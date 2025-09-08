package ledger

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("ledger entry not found")

type Repo interface {
    List(ctx context.Context, filters *ListFilters) ([]LedgerEntry, error)
    Get(ctx context.Context, id int32) (LedgerEntry, error)
    // Create inserts a new ledger entry. If idempotencyKey is provided and a prior
    // matching record exists for the member, it returns that record with replayed=true.
    Create(ctx context.Context, entryType, description string, amount float64, memberID *int32, notes string, idempotencyKey string) (entry LedgerEntry, replayed bool, err error)
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
FROM ledger_entries`
	args := []any{}
	where := ""
	if filters != nil {
		argPos := 0
		if filters.Type != "" {
			argPos++
			if where == "" {
				where = " WHERE"
			} else {
				where += " AND"
			}
			where += " type=$" + itoa(argPos)
			args = append(args, filters.Type)
		}
		if filters.MemberID != nil {
			argPos++
			if where == "" {
				where = " WHERE"
			} else {
				where += " AND"
			}
			where += " member_id=$" + itoa(argPos)
			args = append(args, *filters.MemberID)
		}
	}
	query += where + " ORDER BY id DESC"

	rows, err := r.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []LedgerEntry
	for rows.Next() {
		var e LedgerEntry
		var memberID pgtype.Int4
		var ts pgtype.Timestamptz
		if err := rows.Scan(&e.ID, &e.Type, &e.Amount, &e.Description, &memberID, &e.Notes, &ts); err != nil {
			return nil, err
		}
		if memberID.Valid {
			e.MemberID = &memberID.Int32
		}
		e.CreatedAt = ts.Time
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *PgRepo) Get(ctx context.Context, id int32) (LedgerEntry, error) {
	var e LedgerEntry
	var memberID pgtype.Int4
	err := r.Pool.QueryRow(ctx, `
SELECT id, type, amount, description, member_id, COALESCE(notes,''), created_at
FROM ledger_entries
WHERE id=$1`, id).Scan(&e.ID, &e.Type, &e.Amount, &e.Description, &memberID, &e.Notes, &e.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return LedgerEntry{}, ErrNotFound
		}
		return LedgerEntry{}, err
	}
	if memberID.Valid {
		e.MemberID = &memberID.Int32
	}
	return e, nil
}

func (r *PgRepo) Create(ctx context.Context, entryType, description string, amount float64, memberID *int32, notes string, idempotencyKey string) (LedgerEntry, bool, error) {
    var e LedgerEntry
    var memberIDParam pgtype.Int4
    if memberID != nil {
        memberIDParam.Int32 = *memberID
        memberIDParam.Valid = true
    }

    if idempotencyKey != "" && memberIDParam.Valid {
        // Try insert; if duplicate, select existing by (member_id, idempotency_key)
        err := r.Pool.QueryRow(ctx, `
INSERT INTO ledger_entries (type, amount, description, member_id, notes, idempotency_key)
VALUES ($1,$2,$3,$4,$5,$6)
RETURNING id, type, amount, description, member_id, COALESCE(notes,''), created_at
`, entryType, amount, description, memberIDParam, notes, idempotencyKey).Scan(&e.ID, &e.Type, &e.Amount, &e.Description, &memberIDParam, &e.Notes, &e.CreatedAt)
        if err != nil {
            // On any insert error, attempt to fetch existing idempotent record
            var existing LedgerEntry
            var mid pgtype.Int4
            var ts pgtype.Timestamptz
            err2 := r.Pool.QueryRow(ctx, `
SELECT id, type, amount, description, member_id, COALESCE(notes,''), created_at
FROM ledger_entries
WHERE member_id=$1 AND idempotency_key=$2
LIMIT 1`, memberIDParam, idempotencyKey).Scan(&existing.ID, &existing.Type, &existing.Amount, &existing.Description, &mid, &existing.Notes, &ts)
            if err2 != nil {
                return LedgerEntry{}, false, err
            }
            if mid.Valid {
                existing.MemberID = &mid.Int32
            }
            existing.CreatedAt = ts.Time
            return existing, true, nil
        }
        if memberIDParam.Valid {
            e.MemberID = &memberIDParam.Int32
        }
        return e, false, nil
    }

    err := r.Pool.QueryRow(ctx, `
INSERT INTO ledger_entries (type, amount, description, member_id, notes)
VALUES ($1,$2,$3,$4,$5)
RETURNING id, type, amount, description, member_id, COALESCE(notes,''), created_at
`, entryType, amount, description, memberIDParam, notes).Scan(&e.ID, &e.Type, &e.Amount, &e.Description, &memberIDParam, &e.Notes, &e.CreatedAt)
    if err != nil {
        return LedgerEntry{}, false, err
    }
    if memberIDParam.Valid {
        e.MemberID = &memberIDParam.Int32
    }
    return e, false, nil
}

func itoa(v int) string {
	const digits = "0123456789"
	if v == 0 {
		return "0"
	}
	n := v
	buf := make([]byte, 0, 10)
	for n > 0 {
		buf = append([]byte{digits[n%10]}, buf...)
		n /= 10
	}
	return string(buf)
}
