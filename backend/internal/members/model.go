package members

import "time"

// Member represents a cooperative member/user.
// ID uses BIGINT in DB to allow growth; we expose as int64 here.
type Member struct {
    ID          int64     `json:"id"`
    Email       string    `json:"email"`
    DisplayName string    `json:"display_name"`
    Role        string    `json:"role"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

