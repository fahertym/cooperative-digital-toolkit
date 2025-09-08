package proposals

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("proposal not found")
var ErrConflict = errors.New("invalid state transition")

type Repo interface {
	List(ctx context.Context) ([]Proposal, error)
	Get(ctx context.Context, id int32) (Proposal, error)
	Create(ctx context.Context, title, body string) (Proposal, error)
	Close(ctx context.Context, id int32) (Proposal, error)
}

type PgRepo struct {
	Pool *pgxpool.Pool
}

func NewPgRepo(pool *pgxpool.Pool) *PgRepo {
	return &PgRepo{Pool: pool}
}

func (r *PgRepo) List(ctx context.Context) ([]Proposal, error) {
	rows, err := r.Pool.Query(ctx, `
SELECT id, title, COALESCE(body,''), COALESCE(status,'open'), created_at
FROM proposals
ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Proposal
	for rows.Next() {
		var p Proposal
		var ts pgtype.Timestamptz
		if err := rows.Scan(&p.ID, &p.Title, &p.Body, &p.Status, &ts); err != nil {
			return nil, err
		}
		p.CreatedAt = ts.Time
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *PgRepo) Get(ctx context.Context, id int32) (Proposal, error) {
	var p Proposal
	err := r.Pool.QueryRow(ctx, `
SELECT id, title, COALESCE(body,''), COALESCE(status,'open'), created_at
FROM proposals
WHERE id=$1`, id).Scan(&p.ID, &p.Title, &p.Body, &p.Status, &p.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Proposal{}, ErrNotFound
		}
		return Proposal{}, err
	}
	return p, nil
}

func (r *PgRepo) Create(ctx context.Context, title, body string) (Proposal, error) {
	var p Proposal
	err := r.Pool.QueryRow(ctx, `
INSERT INTO proposals (title, body, status)
VALUES ($1,$2,'open')
RETURNING id, title, COALESCE(body,''), status, created_at
`, title, body).Scan(&p.ID, &p.Title, &p.Body, &p.Status, &p.CreatedAt)
	if err != nil {
		return Proposal{}, err
	}
	return p, nil
}

func (r *PgRepo) Close(ctx context.Context, id int32) (Proposal, error) {
	// Ensure proposal exists and is open
	var current string
	if err := r.Pool.QueryRow(ctx, `SELECT status FROM proposals WHERE id=$1`, id).Scan(&current); err != nil {
		if err == pgx.ErrNoRows {
			return Proposal{}, ErrNotFound
		}
		return Proposal{}, err
	}
	if current != "open" {
		return Proposal{}, ErrConflict
	}

	// Transition to closed
	var p Proposal
	if err := r.Pool.QueryRow(ctx, `
UPDATE proposals
SET status='closed'
WHERE id=$1
RETURNING id, title, COALESCE(body,''), status, created_at
`, id).Scan(&p.ID, &p.Title, &p.Body, &p.Status, &p.CreatedAt); err != nil {
		return Proposal{}, err
	}
	return p, nil
}

// ApplyMigrations creates the proposals table and any necessary schema updates.
func ApplyMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS proposals (
  id SERIAL PRIMARY KEY,
  title TEXT NOT NULL,
  body TEXT,
  status TEXT DEFAULT 'open',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);`)
	if err != nil {
		return err
	}

	// Add status column if it doesn't exist (for existing tables)
	_, err = pool.Exec(ctx, `ALTER TABLE proposals ADD COLUMN IF NOT EXISTS status TEXT DEFAULT 'open';`)
	return err
}
