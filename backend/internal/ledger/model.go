package ledger

import "time"

type LedgerEntry struct {
	ID          int32     `json:"id"`
	Type        string    `json:"type"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	MemberID    *int32    `json:"member_id"`
	Notes       string    `json:"notes"`
	CreatedAt   time.Time `json:"created_at"`
}

// ListFilters holds optional constraints for listing entries.
type ListFilters struct {
    Type     string
    MemberID *int32
    // FromDate, ToDate omitted for now in implementation
    Limit  int
    Offset int
}
