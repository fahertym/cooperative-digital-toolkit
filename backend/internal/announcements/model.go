package announcements

import (
	"time"
)

type Announcement struct {
	ID        int32     `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	AuthorID  *int32    `json:"author_id"` // Optional, null for system announcements
	Priority  string    `json:"priority"`  // 'low', 'normal', 'high', 'urgent'
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AnnouncementWithReadStatus includes read status for a specific member
type AnnouncementWithReadStatus struct {
	Announcement
	IsRead   bool       `json:"is_read"`
	ReadAt   *time.Time `json:"read_at,omitempty"`
	MemberID int32      `json:"member_id"` // The member this read status is for
}

// AnnouncementRead tracks which members have read which announcements
type AnnouncementRead struct {
	AnnouncementID int32     `json:"announcement_id"`
	MemberID       int32     `json:"member_id"`
	ReadAt         time.Time `json:"read_at"`
}
