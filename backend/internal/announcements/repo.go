package announcements

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("announcement not found")

type Repo interface {
	List(ctx context.Context) ([]Announcement, error)
	Get(ctx context.Context, id int32) (Announcement, error)
	Create(ctx context.Context, title, body string, authorID int32, priority string) (Announcement, error)
}

type PgRepo struct {
	Pool *pgxpool.Pool
}

func NewPgRepo(pool *pgxpool.Pool) *PgRepo {
	return &PgRepo{Pool: pool}
}

func (r *PgRepo) List(ctx context.Context) ([]Announcement, error) {
	rows, err := r.Pool.Query(ctx, `
SELECT id, title, body, author_id, COALESCE(priority,'normal'), created_at, updated_at
FROM announcements
ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Announcement
	for rows.Next() {
		var a Announcement
		var createdAt, updatedAt pgtype.Timestamptz
		if err := rows.Scan(&a.ID, &a.Title, &a.Body, &a.AuthorID, &a.Priority, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		a.CreatedAt = createdAt.Time
		a.UpdatedAt = updatedAt.Time
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *PgRepo) Get(ctx context.Context, id int32) (Announcement, error) {
	var a Announcement
	var createdAt, updatedAt pgtype.Timestamptz
	err := r.Pool.QueryRow(ctx, `
SELECT id, title, body, author_id, COALESCE(priority,'normal'), created_at, updated_at
FROM announcements
WHERE id=$1`, id).Scan(&a.ID, &a.Title, &a.Body, &a.AuthorID, &a.Priority, &createdAt, &updatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Announcement{}, ErrNotFound
		}
		return Announcement{}, err
	}
	a.CreatedAt = createdAt.Time
	a.UpdatedAt = updatedAt.Time
	return a, nil
}

func (r *PgRepo) Create(ctx context.Context, title, body string, authorID int32, priority string) (Announcement, error) {
	var a Announcement
	var createdAt, updatedAt pgtype.Timestamptz
	err := r.Pool.QueryRow(ctx, `
INSERT INTO announcements (title, body, author_id, priority)
VALUES ($1,$2,$3,$4)
RETURNING id, title, body, author_id, COALESCE(priority,'normal'), created_at, updated_at
`, title, body, authorID, priority).Scan(&a.ID, &a.Title, &a.Body, &a.AuthorID, &a.Priority, &createdAt, &updatedAt)
	if err != nil {
		return Announcement{}, err
	}
	a.CreatedAt = createdAt.Time
	a.UpdatedAt = updatedAt.Time
	return a, nil
}

// ApplyMigrations creates the announcements table and any necessary schema updates.
func ApplyMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS announcements (
  id SERIAL PRIMARY KEY,
  title TEXT NOT NULL,
  body TEXT NOT NULL,
  author_id INTEGER NOT NULL,
  priority TEXT DEFAULT 'normal',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);`)
	return err
}
