package announcements

import "time"

type Announcement struct {
	ID        int32     `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	AuthorID  int32     `json:"author_id"`
	Priority  string    `json:"priority"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
