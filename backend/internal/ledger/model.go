package ledger

import "time"

type Entry struct {
	ID          int32     `json:"id"`
	Type        string    `json:"type"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	MemberID    *int32    `json:"member_id"`
	Notes       string    `json:"notes"`
	CreatedAt   time.Time `json:"created_at"`
}
