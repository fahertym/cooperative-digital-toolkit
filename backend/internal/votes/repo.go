package votes

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound       = errors.New("vote not found")
	ErrAlreadyVoted   = errors.New("member already voted on this proposal")
	ErrProposalClosed = errors.New("proposal is closed")
)

type Repo interface {
    List(ctx context.Context, proposalID int32, limit, offset int) ([]Vote, error)
    Get(ctx context.Context, proposalID, memberID int32) (Vote, error)
    Create(ctx context.Context, proposalID, memberID int32, choice, notes string) (Vote, error)
    Update(ctx context.Context, proposalID, memberID int32, choice, notes string) (Vote, error)
    GetTally(ctx context.Context, proposalID int32) (Tally, error)
}

type PgRepo struct {
	Pool *pgxpool.Pool
}

func NewPgRepo(pool *pgxpool.Pool) *PgRepo {
	return &PgRepo{Pool: pool}
}

func (r *PgRepo) List(ctx context.Context, proposalID int32, limit, offset int) ([]Vote, error) {
    query := `
SELECT id, proposal_id, member_id, choice, COALESCE(notes,''), created_at
FROM votes
WHERE proposal_id=$1
ORDER BY created_at ASC`
    args := []any{proposalID}
    if limit > 0 {
        query += ` LIMIT $2`
        args = append(args, limit)
        if offset > 0 {
            query += ` OFFSET $3`
            args = append(args, offset)
        }
    } else if offset > 0 {
        query += ` OFFSET $2`
        args = append(args, offset)
    }
    rows, err := r.Pool.Query(ctx, query, args...)
    if err != nil {
        return nil, err
    }
	defer rows.Close()

	var out []Vote
	for rows.Next() {
		var v Vote
		var ts pgtype.Timestamptz
		if err := rows.Scan(&v.ID, &v.ProposalID, &v.MemberID, &v.Choice, &v.Notes, &ts); err != nil {
			return nil, err
		}
		v.CreatedAt = ts.Time
		out = append(out, v)
	}
	return out, rows.Err()
}

func (r *PgRepo) Get(ctx context.Context, proposalID, memberID int32) (Vote, error) {
	var v Vote
	var ts pgtype.Timestamptz
	err := r.Pool.QueryRow(ctx, `
SELECT id, proposal_id, member_id, choice, COALESCE(notes,''), created_at
FROM votes
WHERE proposal_id=$1 AND member_id=$2`, proposalID, memberID).Scan(&v.ID, &v.ProposalID, &v.MemberID, &v.Choice, &v.Notes, &ts)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Vote{}, ErrNotFound
		}
		return Vote{}, err
	}
	v.CreatedAt = ts.Time
	return v, nil
}

func (r *PgRepo) Create(ctx context.Context, proposalID, memberID int32, choice, notes string) (Vote, error) {
	// Check if proposal is open
	var status string
	if err := r.Pool.QueryRow(ctx, `SELECT status FROM proposals WHERE id=$1`, proposalID).Scan(&status); err != nil {
		if err == pgx.ErrNoRows {
			return Vote{}, ErrNotFound
		}
		return Vote{}, err
	}
	if status != "open" {
		return Vote{}, ErrProposalClosed
	}

	// Check if member already voted
	var exists bool
	if err := r.Pool.QueryRow(ctx, `SELECT true FROM votes WHERE proposal_id=$1 AND member_id=$2`, proposalID, memberID).Scan(&exists); err == nil {
		return Vote{}, ErrAlreadyVoted
	}

	var v Vote
	var ts pgtype.Timestamptz
	err := r.Pool.QueryRow(ctx, `
INSERT INTO votes (proposal_id, member_id, choice, notes)
VALUES ($1,$2,$3,$4)
RETURNING id, proposal_id, member_id, choice, COALESCE(notes,''), created_at
`, proposalID, memberID, choice, notes).Scan(&v.ID, &v.ProposalID, &v.MemberID, &v.Choice, &v.Notes, &ts)
	if err != nil {
		return Vote{}, err
	}
	v.CreatedAt = ts.Time
	return v, nil
}

func (r *PgRepo) Update(ctx context.Context, proposalID, memberID int32, choice, notes string) (Vote, error) {
	// Check if proposal is open
	var status string
	if err := r.Pool.QueryRow(ctx, `SELECT status FROM proposals WHERE id=$1`, proposalID).Scan(&status); err != nil {
		if err == pgx.ErrNoRows {
			return Vote{}, ErrNotFound
		}
		return Vote{}, err
	}
	if status != "open" {
		return Vote{}, ErrProposalClosed
	}

	var v Vote
	var ts pgtype.Timestamptz
	err := r.Pool.QueryRow(ctx, `
UPDATE votes
SET choice=$3, notes=$4
WHERE proposal_id=$1 AND member_id=$2
RETURNING id, proposal_id, member_id, choice, COALESCE(notes,''), created_at
`, proposalID, memberID, choice, notes).Scan(&v.ID, &v.ProposalID, &v.MemberID, &v.Choice, &v.Notes, &ts)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Vote{}, ErrNotFound
		}
		return Vote{}, err
	}
	v.CreatedAt = ts.Time
	return v, nil
}

func (r *PgRepo) GetTally(ctx context.Context, proposalID int32) (Tally, error) {
	// Get proposal status
	var status string
	if err := r.Pool.QueryRow(ctx, `SELECT status FROM proposals WHERE id=$1`, proposalID).Scan(&status); err != nil {
		if err == pgx.ErrNoRows {
			return Tally{}, ErrNotFound
		}
		return Tally{}, err
	}

	// For now, assume 10 eligible members (in real system, this would come from members table)
	totalEligible := 10

	// Count votes by choice
	rows, err := r.Pool.Query(ctx, `
SELECT choice, COUNT(*) as count
FROM votes
WHERE proposal_id=$1
GROUP BY choice`, proposalID)
	if err != nil {
		return Tally{}, err
	}
	defer rows.Close()

	results := map[string]int{"for": 0, "against": 0, "abstain": 0}
	votesCast := 0

	for rows.Next() {
		var choice string
		var count int
		if err := rows.Scan(&choice, &count); err != nil {
			return Tally{}, err
		}
		results[choice] = count
		votesCast += count
	}

	// Calculate quorum (simple majority of eligible members)
	quorumMet := votesCast >= (totalEligible/2 + 1)

	// Determine outcome
	var outcome string
	if status == "open" {
		outcome = "pending"
	} else {
		if results["for"] > results["against"] {
			outcome = "passed"
		} else {
			outcome = "failed"
		}
	}

	return Tally{
		ProposalID:    proposalID,
		Status:        status,
		TotalEligible: totalEligible,
		VotesCast:     votesCast,
		QuorumMet:     quorumMet,
		Results:       results,
		Outcome:       outcome,
	}, nil
}
