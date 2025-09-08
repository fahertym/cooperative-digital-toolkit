package ledger

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("ledger entry not found")

type Repo interface {
	List(ctx context.Context, filters *ListFilters) ([]LedgerEntry, error)
	Get(ctx context.Context, id int32) (LedgerEntry, error)
	Create(ctx context.Context, entryType, description string, amount float64, memberID *int32, notes string) (LedgerEntry, error)
}

type ListFilters struct {
	Type     string
	MemberID *int32
	FromDate *string // RFC3339 format
	ToDate   *string // RFC3339 format
}

type PgRepo struct {
	Pool *pgxpool.Pool
}

func NewPgRepo(pool *pgxpool.Pool) *PgRepo {
	return &PgRepo{Pool: pool}
}

func (r *PgRepo) List(ctx context.Context, filters *ListFilters) ([]LedgerEntry, error) {
	query := `
SELECT id, type, amount, description, member_id, COALESCE(notes,''), created_at
FROM ledger_entries
WHERE 1=1`

	args := []interface{}{}
	argPos := 1

	// Add filters
	if filters != nil {
		if filters.Type != "" {
			query += ` AND type = $` + fmt.Sprintf("%d", argPos)
			args = append(args, filters.Type)
			argPos++
		}
		if filters.MemberID != nil {
			query += ` AND member_id = $` + fmt.Sprintf("%d", argPos)
			args = append(args, *filters.MemberID)
			argPos++
		}
		if filters.FromDate != nil {
			query += ` AND created_at >= $` + fmt.Sprintf("%d", argPos)
			args = append(args, *filters.FromDate)
			argPos++
		}
		if filters.ToDate != nil {
			query += ` AND created_at <= $` + fmt.Sprintf("%d", argPos)
			args = append(args, *filters.ToDate)
			argPos++
		}
	}

	query += `
ORDER BY created_at DESC`

	rows, err := r.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []LedgerEntry
	for rows.Next() {
		var entry LedgerEntry
		var ts pgtype.Timestamptz
		if err := rows.Scan(&entry.ID, &entry.Type, &entry.Amount, &entry.Description, &entry.MemberID, &entry.Notes, &ts); err != nil {
			return nil, err
		}
		entry.CreatedAt = ts.Time
		out = append(out, entry)
	}
	return out, rows.Err()
}

func (r *PgRepo) Get(ctx context.Context, id int32) (LedgerEntry, error) {
	var entry LedgerEntry
	var ts pgtype.Timestamptz
	err := r.Pool.QueryRow(ctx, `
SELECT id, type, amount, description, member_id, COALESCE(notes,''), created_at
FROM ledger_entries
WHERE id=$1`, id).Scan(&entry.ID, &entry.Type, &entry.Amount, &entry.Description, &entry.MemberID, &entry.Notes, &ts)
	if err != nil {
		if err == pgx.ErrNoRows {
			return LedgerEntry{}, ErrNotFound
		}
		return LedgerEntry{}, err
	}
	entry.CreatedAt = ts.Time
	return entry, nil
}

func (r *PgRepo) Create(ctx context.Context, entryType, description string, amount float64, memberID *int32, notes string) (LedgerEntry, error) {
	var entry LedgerEntry
	var ts pgtype.Timestamptz
	err := r.Pool.QueryRow(ctx, `
INSERT INTO ledger_entries (type, amount, description, member_id, notes)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, type, amount, description, member_id, COALESCE(notes,''), created_at
`, entryType, amount, description, memberID, notes).Scan(&entry.ID, &entry.Type, &entry.Amount, &entry.Description, &entry.MemberID, &entry.Notes, &ts)
	if err != nil {
		return LedgerEntry{}, err
	}
	entry.CreatedAt = ts.Time
	return entry, nil
}
