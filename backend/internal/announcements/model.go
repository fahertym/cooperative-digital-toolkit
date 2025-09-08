package announcements

import "time"

type Announcement struct {
    ID        int32     `json:"id"`
    Title     string    `json:"title"`
    Body      string    `json:"body"`
    AuthorID  *int64    `json:"author_id"`
    Priority  string    `json:"priority"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// AnnouncementRead represents a record that a given member read a given announcement.
type AnnouncementRead struct {
    AnnouncementID int32     `json:"announcement_id"`
    MemberID       int64     `json:"member_id"`
    ReadAt         time.Time `json:"read_at"`
}

// AnnouncementWithReadStatus augments an announcement with per-member read info.
type AnnouncementWithReadStatus struct {
    Announcement
    MemberID int64      `json:"member_id,omitempty"`
    IsRead   bool       `json:"is_read"`
    ReadAt   *time.Time `json:"read_at,omitempty"`
}

// ListFilters constrains which announcements are returned.
type ListFilters struct {
    Priority   string
    AuthorID   *int64
    OnlyUnread bool
    Limit      int
    Offset     int
}
