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
	List(ctx context.Context) ([]Entry, error)
	Get(ctx context.Context, id int32) (Entry, error)
	Create(ctx context.Context, entryType string, amount float64, description string, memberID *int32, notes string) (Entry, error)
}

type PgRepo struct {
	Pool *pgxpool.Pool
}

func NewPgRepo(pool *pgxpool.Pool) *PgRepo {
	return &PgRepo{Pool: pool}
}

func (r *PgRepo) List(ctx context.Context) ([]Entry, error) {
	rows, err := r.Pool.Query(ctx, `
SELECT id, type, amount, description, member_id, COALESCE(notes,''), created_at
FROM ledger_entries
ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Entry
	for rows.Next() {
		var e Entry
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

func (r *PgRepo) Get(ctx context.Context, id int32) (Entry, error) {
	var e Entry
	var memberID pgtype.Int4
	err := r.Pool.QueryRow(ctx, `
SELECT id, type, amount, description, member_id, COALESCE(notes,''), created_at
FROM ledger_entries
WHERE id=$1`, id).Scan(&e.ID, &e.Type, &e.Amount, &e.Description, &memberID, &e.Notes, &e.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Entry{}, ErrNotFound
		}
		return Entry{}, err
	}
	if memberID.Valid {
		e.MemberID = &memberID.Int32
	}
	return e, nil
}

func (r *PgRepo) Create(ctx context.Context, entryType string, amount float64, description string, memberID *int32, notes string) (Entry, error) {
	var e Entry
	var memberIDParam pgtype.Int4
	if memberID != nil {
		memberIDParam.Int32 = *memberID
		memberIDParam.Valid = true
	}

	err := r.Pool.QueryRow(ctx, `
INSERT INTO ledger_entries (type, amount, description, member_id, notes)
VALUES ($1,$2,$3,$4,$5)
RETURNING id, type, amount, description, member_id, COALESCE(notes,''), created_at
`, entryType, amount, description, memberIDParam, notes).Scan(&e.ID, &e.Type, &e.Amount, &e.Description, &memberIDParam, &e.Notes, &e.CreatedAt)
	if err != nil {
		return Entry{}, err
	}
	if memberIDParam.Valid {
		e.MemberID = &memberIDParam.Int32
	}
	return e, nil
}

// ApplyMigrations creates the ledger_entries table and any necessary schema updates.
func ApplyMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS ledger_entries (
  id SERIAL PRIMARY KEY,
  type TEXT NOT NULL,
  amount DECIMAL(10,2) NOT NULL,
  description TEXT NOT NULL,
  member_id INTEGER,
  notes TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);`)
	return err
}
