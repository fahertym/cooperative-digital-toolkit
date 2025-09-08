package members

import (
    "context"
    "errors"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgconn"
    "github.com/jackc/pgx/v5/pgtype"
    "github.com/jackc/pgx/v5/pgxpool"
)

var (
    ErrNotFound = errors.New("member not found")
    ErrConflict = errors.New("member conflict")
)

type Repo interface {
    Create(ctx context.Context, email, displayName, role string) (Member, error)
    GetByID(ctx context.Context, id int64) (Member, error)
    GetByEmail(ctx context.Context, email string) (Member, error)
}

type PgRepo struct{ Pool *pgxpool.Pool }

func NewPgRepo(pool *pgxpool.Pool) *PgRepo { return &PgRepo{Pool: pool} }

func (r *PgRepo) Create(ctx context.Context, email, displayName, role string) (Member, error) {
    if role == "" { role = "member" }
    var m Member
    var createdAt, updatedAt pgtype.Timestamptz
    err := r.Pool.QueryRow(ctx, `
INSERT INTO members (email, display_name, role)
VALUES ($1,$2,$3)
RETURNING id, email, display_name, role, created_at, updated_at
`, email, displayName, role).Scan(
        &m.ID, &m.Email, &m.DisplayName, &m.Role, &createdAt, &updatedAt,
    )
    if err != nil {
        var pgErr *pgconn.PgError
        if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
            return Member{}, ErrConflict
        }
        return Member{}, err
    }
    m.CreatedAt = createdAt.Time
    m.UpdatedAt = updatedAt.Time
    return m, nil
}

func (r *PgRepo) GetByID(ctx context.Context, id int64) (Member, error) {
    var m Member
    var createdAt, updatedAt pgtype.Timestamptz
    err := r.Pool.QueryRow(ctx, `
SELECT id, email, display_name, role, created_at, updated_at
FROM members WHERE id=$1
`, id).Scan(&m.ID, &m.Email, &m.DisplayName, &m.Role, &createdAt, &updatedAt)
    if err != nil {
        if err == pgx.ErrNoRows { return Member{}, ErrNotFound }
        return Member{}, err
    }
    m.CreatedAt = createdAt.Time
    m.UpdatedAt = updatedAt.Time
    return m, nil
}

func (r *PgRepo) GetByEmail(ctx context.Context, email string) (Member, error) {
    var m Member
    var createdAt, updatedAt pgtype.Timestamptz
    err := r.Pool.QueryRow(ctx, `
SELECT id, email, display_name, role, created_at, updated_at
FROM members WHERE email=$1
`, email).Scan(&m.ID, &m.Email, &m.DisplayName, &m.Role, &createdAt, &updatedAt)
    if err != nil {
        if err == pgx.ErrNoRows { return Member{}, ErrNotFound }
        return Member{}, err
    }
    m.CreatedAt = createdAt.Time
    m.UpdatedAt = updatedAt.Time
    return m, nil
}

