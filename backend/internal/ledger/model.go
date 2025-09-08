package ledger

import (
	"time"
)

type LedgerEntry struct {
	ID          int32     `json:"id"`
	Type        string    `json:"type"`        // 'dues', 'contribution', 'expense', 'income'
	Amount      float64   `json:"amount"`      // decimal amount in dollars
	Description string    `json:"description"` // required description
	MemberID    *int32    `json:"member_id"`   // optional, null for org-level entries
	Notes       string    `json:"notes"`       // optional additional notes
	CreatedAt   time.Time `json:"created_at"`
}
