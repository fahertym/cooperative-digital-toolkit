package announcements

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("announcement not found")

type Repo interface {
	List(ctx context.Context, memberID *int32, filters *ListFilters) ([]AnnouncementWithReadStatus, error)
	Get(ctx context.Context, id int32, memberID *int32) (AnnouncementWithReadStatus, error)
	Create(ctx context.Context, title, body string, authorID *int32, priority string) (Announcement, error)
	MarkAsRead(ctx context.Context, announcementID, memberID int32) error
	GetUnreadCount(ctx context.Context, memberID int32) (int, error)
}

type ListFilters struct {
	Priority   string
	AuthorID   *int32
	OnlyUnread bool
	FromDate   *string // RFC3339 format
	ToDate     *string // RFC3339 format
}

type PgRepo struct {
	Pool *pgxpool.Pool
}

func NewPgRepo(pool *pgxpool.Pool) *PgRepo {
	return &PgRepo{Pool: pool}
}

func (r *PgRepo) List(ctx context.Context, memberID *int32, filters *ListFilters) ([]AnnouncementWithReadStatus, error) {
	query := `
SELECT 
  a.id, a.title, a.body, a.author_id, a.priority, a.created_at, a.updated_at,
  CASE WHEN ar.announcement_id IS NOT NULL THEN true ELSE false END as is_read,
  ar.read_at
FROM announcements a
LEFT JOIN announcement_reads ar ON a.id = ar.announcement_id AND ar.member_id = $1
WHERE 1=1`

	args := []interface{}{}
	argPos := 1

	// Always include memberID for read status (can be null)
	args = append(args, memberID)
	argPos++

	// Add filters
	if filters != nil {
		if filters.Priority != "" {
			query += ` AND a.priority = $` + fmt.Sprintf("%d", argPos)
			args = append(args, filters.Priority)
			argPos++
		}
		if filters.AuthorID != nil {
			query += ` AND a.author_id = $` + fmt.Sprintf("%d", argPos)
			args = append(args, *filters.AuthorID)
			argPos++
		}
		if filters.OnlyUnread && memberID != nil {
			query += ` AND ar.announcement_id IS NULL`
		}
		if filters.FromDate != nil {
			query += ` AND a.created_at >= $` + fmt.Sprintf("%d", argPos)
			args = append(args, *filters.FromDate)
			argPos++
		}
		if filters.ToDate != nil {
			query += ` AND a.created_at <= $` + fmt.Sprintf("%d", argPos)
			args = append(args, *filters.ToDate)
			argPos++
		}
	}

	query += `
ORDER BY a.created_at DESC`

	rows, err := r.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []AnnouncementWithReadStatus
	for rows.Next() {
		var item AnnouncementWithReadStatus
		var createdAt, updatedAt pgtype.Timestamptz
		var readAt pgtype.Timestamptz

		if err := rows.Scan(
			&item.ID, &item.Title, &item.Body, &item.AuthorID, &item.Priority,
			&createdAt, &updatedAt, &item.IsRead, &readAt,
		); err != nil {
			return nil, err
		}

		item.CreatedAt = createdAt.Time
		item.UpdatedAt = updatedAt.Time
		if readAt.Valid {
			item.ReadAt = &readAt.Time
		}
		if memberID != nil {
			item.MemberID = *memberID
		}

		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *PgRepo) Get(ctx context.Context, id int32, memberID *int32) (AnnouncementWithReadStatus, error) {
	query := `
SELECT 
  a.id, a.title, a.body, a.author_id, a.priority, a.created_at, a.updated_at,
  CASE WHEN ar.announcement_id IS NOT NULL THEN true ELSE false END as is_read,
  ar.read_at
FROM announcements a
LEFT JOIN announcement_reads ar ON a.id = ar.announcement_id AND ar.member_id = $2
WHERE a.id = $1`

	var item AnnouncementWithReadStatus
	var createdAt, updatedAt pgtype.Timestamptz
	var readAt pgtype.Timestamptz

	err := r.Pool.QueryRow(ctx, query, id, memberID).Scan(
		&item.ID, &item.Title, &item.Body, &item.AuthorID, &item.Priority,
		&createdAt, &updatedAt, &item.IsRead, &readAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return AnnouncementWithReadStatus{}, ErrNotFound
		}
		return AnnouncementWithReadStatus{}, err
	}

	item.CreatedAt = createdAt.Time
	item.UpdatedAt = updatedAt.Time
	if readAt.Valid {
		item.ReadAt = &readAt.Time
	}
	if memberID != nil {
		item.MemberID = *memberID
	}

	return item, nil
}

func (r *PgRepo) Create(ctx context.Context, title, body string, authorID *int32, priority string) (Announcement, error) {
	var announcement Announcement
	var createdAt, updatedAt pgtype.Timestamptz

	err := r.Pool.QueryRow(ctx, `
INSERT INTO announcements (title, body, author_id, priority)
VALUES ($1, $2, $3, $4)
RETURNING id, title, body, author_id, priority, created_at, updated_at
`, title, body, authorID, priority).Scan(
		&announcement.ID, &announcement.Title, &announcement.Body,
		&announcement.AuthorID, &announcement.Priority, &createdAt, &updatedAt,
	)
	if err != nil {
		return Announcement{}, err
	}

	announcement.CreatedAt = createdAt.Time
	announcement.UpdatedAt = updatedAt.Time
	return announcement, nil
}

func (r *PgRepo) MarkAsRead(ctx context.Context, announcementID, memberID int32) error {
	// Use UPSERT to handle duplicate reads gracefully
	_, err := r.Pool.Exec(ctx, `
INSERT INTO announcement_reads (announcement_id, member_id)
VALUES ($1, $2)
ON CONFLICT (announcement_id, member_id) DO NOTHING
`, announcementID, memberID)
	return err
}

func (r *PgRepo) GetUnreadCount(ctx context.Context, memberID int32) (int, error) {
	var count int
	err := r.Pool.QueryRow(ctx, `
SELECT COUNT(*)
FROM announcements a
LEFT JOIN announcement_reads ar ON a.id = ar.announcement_id AND ar.member_id = $1
WHERE ar.announcement_id IS NULL
`, memberID).Scan(&count)
	return count, err
}
