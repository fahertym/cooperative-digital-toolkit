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
    List(ctx context.Context, memberID *int64, filters *ListFilters) ([]AnnouncementWithReadStatus, error)
    Get(ctx context.Context, id int32, memberID *int64) (AnnouncementWithReadStatus, error)
    Create(ctx context.Context, title, body string, authorID *int64, priority string) (Announcement, error)
    MarkAsRead(ctx context.Context, announcementID int32, memberID int64) error
    GetUnreadCount(ctx context.Context, memberID int64) (int, error)
}

type PgRepo struct {
	Pool *pgxpool.Pool
}

func NewPgRepo(pool *pgxpool.Pool) *PgRepo {
	return &PgRepo{Pool: pool}
}

func (r *PgRepo) List(ctx context.Context, memberID *int64, filters *ListFilters) ([]AnnouncementWithReadStatus, error) {
    query := `
SELECT a.id, a.title, a.body, a.author_id, COALESCE(a.priority,'normal'), a.created_at, a.updated_at,
       (ar.read_at IS NOT NULL) AS is_read, ar.read_at
FROM announcements a`
    args := []any{}
    if memberID != nil {
        query += ` LEFT JOIN announcement_reads ar ON ar.announcement_id=a.id AND ar.member_id=$1`
        args = append(args, *memberID)
    } else {
        query += ` LEFT JOIN LATERAL (SELECT NULL::TIMESTAMPTZ AS read_at) ar ON true`
    }

    where := ""
    argPos := len(args)
    if filters != nil {
        if filters.Priority != "" {
            if where == "" {
                where = " WHERE"
            } else {
                where += " AND"
            }
            argPos++
            where += " a.priority=$" + itoa(argPos)
            args = append(args, filters.Priority)
        }
        if filters.AuthorID != nil {
            if where == "" {
                where = " WHERE"
            } else {
                where += " AND"
            }
            argPos++
            where += " a.author_id=$" + itoa(argPos)
            args = append(args, *filters.AuthorID)
        }
        if filters.OnlyUnread && memberID != nil {
            if where == "" {
                where = " WHERE"
            } else {
                where += " AND"
            }
            where += " ar.read_at IS NULL"
        }
    }
    query += where + " ORDER BY a.id DESC"

    // Pagination
    if filters != nil {
        if filters.Limit > 0 {
            argPos++
            query += " LIMIT $" + itoa(argPos)
            args = append(args, filters.Limit)
        }
        if filters.Offset > 0 {
            argPos++
            query += " OFFSET $" + itoa(argPos)
            args = append(args, filters.Offset)
        }
    }

	rows, err := r.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

    var out []AnnouncementWithReadStatus
    for rows.Next() {
        var a AnnouncementWithReadStatus
        var createdAt, updatedAt pgtype.Timestamptz
        var authorID pgtype.Int8
        var readAt pgtype.Timestamptz
        if err := rows.Scan(&a.ID, &a.Title, &a.Body, &authorID, &a.Priority, &createdAt, &updatedAt, &a.IsRead, &readAt); err != nil {
            return nil, err
        }
        if authorID.Valid {
            a.AuthorID = &authorID.Int64
        }
        a.CreatedAt = createdAt.Time
        a.UpdatedAt = updatedAt.Time
        if memberID != nil {
            a.MemberID = *memberID
        }
        if readAt.Valid {
            t := readAt.Time
            a.ReadAt = &t
        }
        out = append(out, a)
    }
    return out, rows.Err()
}

func (r *PgRepo) Get(ctx context.Context, id int32, memberID *int64) (AnnouncementWithReadStatus, error) {
	query := `
SELECT a.id, a.title, a.body, a.author_id, COALESCE(a.priority,'normal'), a.created_at, a.updated_at,
       (ar.read_at IS NOT NULL) AS is_read, ar.read_at
FROM announcements a`
	args := []any{id}
    if memberID != nil {
        query += ` LEFT JOIN announcement_reads ar ON ar.announcement_id=a.id AND ar.member_id=$2`
        args = append(args, *memberID)
    } else {
        query += ` LEFT JOIN LATERAL (SELECT NULL::TIMESTAMPTZ AS read_at) ar ON true`
    }
	query += ` WHERE a.id=$1`

    var a AnnouncementWithReadStatus
    var createdAt, updatedAt pgtype.Timestamptz
    var authorID pgtype.Int8
    var readAt pgtype.Timestamptz
    err := r.Pool.QueryRow(ctx, query, args...).Scan(&a.ID, &a.Title, &a.Body, &authorID, &a.Priority, &createdAt, &updatedAt, &a.IsRead, &readAt)
    if err != nil {
        if err == pgx.ErrNoRows {
            return AnnouncementWithReadStatus{}, ErrNotFound
        }
        return AnnouncementWithReadStatus{}, err
    }
    if authorID.Valid {
        a.AuthorID = &authorID.Int64
    }
    a.CreatedAt = createdAt.Time
    a.UpdatedAt = updatedAt.Time
    if memberID != nil {
        a.MemberID = *memberID
    }
    if readAt.Valid {
        t := readAt.Time
        a.ReadAt = &t
    }
    return a, nil
}

func (r *PgRepo) Create(ctx context.Context, title, body string, authorID *int64, priority string) (Announcement, error) {
    var a Announcement
    var createdAt, updatedAt pgtype.Timestamptz
    var authorParam pgtype.Int8
    if authorID != nil {
        authorParam.Int64 = *authorID
        authorParam.Valid = true
    }
    err := r.Pool.QueryRow(ctx, `
INSERT INTO announcements (title, body, author_id, priority)
VALUES ($1,$2,$3,$4)
RETURNING id, title, body, author_id, COALESCE(priority,'normal'), created_at, updated_at
`, title, body, authorParam, priority).Scan(&a.ID, &a.Title, &a.Body, &authorParam, &a.Priority, &createdAt, &updatedAt)
    if err != nil {
        return Announcement{}, err
    }
    if authorParam.Valid {
        a.AuthorID = &authorParam.Int64
    }
    a.CreatedAt = createdAt.Time
    a.UpdatedAt = updatedAt.Time
    return a, nil
}

func (r *PgRepo) MarkAsRead(ctx context.Context, announcementID int32, memberID int64) error {
    var exists bool
    if err := r.Pool.QueryRow(ctx, `SELECT true FROM announcements WHERE id=$1`, announcementID).Scan(&exists); err != nil {
        if err == pgx.ErrNoRows {
            return ErrNotFound
        }
        return err
    }
    _, err := r.Pool.Exec(ctx, `
INSERT INTO announcement_reads (announcement_id, member_id)
VALUES ($1,$2)
ON CONFLICT (announcement_id, member_id) DO NOTHING`, announcementID, memberID)
    return err
}

func (r *PgRepo) GetUnreadCount(ctx context.Context, memberID int64) (int, error) {
    var count int
    err := r.Pool.QueryRow(ctx, `
SELECT COUNT(*)
FROM announcements a
LEFT JOIN announcement_reads ar
  ON ar.announcement_id=a.id AND ar.member_id=$1
WHERE ar.announcement_id IS NULL`, memberID).Scan(&count)
    if err != nil {
        return 0, err
    }
    return count, nil
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
