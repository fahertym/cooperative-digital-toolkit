package votes

import "time"

type Vote struct {
	ID         int32     `json:"id"`
	ProposalID int32     `json:"proposal_id"`
	MemberID   int32     `json:"member_id"`
	Choice     string    `json:"choice"` // "for", "against", "abstain"
	Notes      string    `json:"notes"`
	CreatedAt  time.Time `json:"created_at"`
}

type Tally struct {
	ProposalID    int32          `json:"proposal_id"`
	Status        string         `json:"status"` // "open", "closed"
	TotalEligible int            `json:"total_eligible"`
	VotesCast     int            `json:"votes_cast"`
	QuorumMet     bool           `json:"quorum_met"`
	Results       map[string]int `json:"results"`
	Outcome       string         `json:"outcome"` // "passed", "failed", "pending"
}
