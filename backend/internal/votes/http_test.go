package votes

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

// ---- Mock Repo ----

type mockRepo struct {
	votes     []Vote
	nextID    int32
	statusFor map[int32]string // proposal_id -> status ("open"/"closed")
}

func (m *mockRepo) ensureInit() {
	if m.statusFor == nil {
		m.statusFor = map[int32]string{}
	}
}

func (m *mockRepo) List(_ context.Context, proposalID int32) ([]Vote, error) {
	var out []Vote
	for _, v := range m.votes {
		if v.ProposalID == proposalID {
			out = append(out, v)
		}
	}
	return out, nil
}

func (m *mockRepo) Get(_ context.Context, proposalID, memberID int32) (Vote, error) {
	for _, v := range m.votes {
		if v.ProposalID == proposalID && v.MemberID == memberID {
			return v, nil
		}
	}
	return Vote{}, ErrNotFound
}

func (m *mockRepo) Create(_ context.Context, proposalID, memberID int32, choice, notes string) (Vote, error) {
	m.ensureInit()
	if m.statusFor[proposalID] == "" {
		m.statusFor[proposalID] = "open"
	}
	if m.statusFor[proposalID] != "open" {
		return Vote{}, ErrProposalClosed
	}
	// duplicate?
	for _, v := range m.votes {
		if v.ProposalID == proposalID && v.MemberID == memberID {
			return Vote{}, ErrAlreadyVoted
		}
	}
	if m.nextID == 0 {
		m.nextID = 1
	}
	v := Vote{ID: m.nextID, ProposalID: proposalID, MemberID: memberID, Choice: choice, Notes: notes}
	m.nextID++
	m.votes = append(m.votes, v)
	return v, nil
}

func (m *mockRepo) Update(_ context.Context, proposalID, memberID int32, choice, notes string) (Vote, error) {
	m.ensureInit()
	if m.statusFor[proposalID] != "open" {
		return Vote{}, ErrProposalClosed
	}
	for i, v := range m.votes {
		if v.ProposalID == proposalID && v.MemberID == memberID {
			v.Choice = choice
			v.Notes = notes
			m.votes[i] = v
			return v, nil
		}
	}
	return Vote{}, ErrNotFound
}

func (m *mockRepo) GetTally(_ context.Context, proposalID int32) (Tally, error) {
	m.ensureInit()
	if _, ok := m.statusFor[proposalID]; !ok {
		return Tally{}, ErrNotFound
	}
	results := map[string]int{"for": 0, "against": 0, "abstain": 0}
	votesCast := 0
	for _, v := range m.votes {
		if v.ProposalID == proposalID {
			results[v.Choice]++
			votesCast++
		}
	}
	totalEligible := 10
	quorumMet := votesCast >= (totalEligible/2 + 1)
	outcome := "pending"
	if m.statusFor[proposalID] != "open" {
		if results["for"] > results["against"] {
			outcome = "passed"
		} else {
			outcome = "failed"
		}
	}
	return Tally{
		ProposalID:    proposalID,
		Status:        m.statusFor[proposalID],
		TotalEligible: totalEligible,
		VotesCast:     votesCast,
		QuorumMet:     quorumMet,
		Results:       results,
		Outcome:       outcome,
	}, nil
}

// ---- Test Router Setup ----

func testRouter(repo Repo) http.Handler {
	r := chi.NewRouter()
	h := Handlers{Repo: repo}
	r.Route("/api", func(api chi.Router) { Mount(api, h) })
	return r
}

// ---- Tests ----

func TestVotes_CreateAndList(t *testing.T) {
	repo := &mockRepo{}
	r := testRouter(repo)

	// Create vote
	payload := `{"choice":"for"}`
	req := httptest.NewRequest("POST", "/api/proposals/42/votes", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Id", "1")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create: want 201 got %d (%s)", rr.Code, rr.Body.String())
	}

	// List
	reqL := httptest.NewRequest("GET", "/api/proposals/42/votes", nil)
	rrL := httptest.NewRecorder()
	r.ServeHTTP(rrL, reqL)
	if rrL.Code != http.StatusOK {
		t.Fatalf("list: want 200 got %d", rrL.Code)
	}
	var list []Vote
	_ = json.Unmarshal(rrL.Body.Bytes(), &list)
	if len(list) != 1 || list[0].MemberID != 1 || list[0].Choice != "for" {
		t.Fatalf("unexpected list: %+v", list)
	}
}

func TestVotes_InvalidChoiceAndDuplicate(t *testing.T) {
	repo := &mockRepo{}
	r := testRouter(repo)

	// Invalid choice
	bad := `{"choice":"maybe"}`
	req := httptest.NewRequest("POST", "/api/proposals/7/votes", strings.NewReader(bad))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Id", "1")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("invalid choice: want 400 got %d", rr.Code)
	}

	// Create valid
	good := `{"choice":"against"}`
	req2 := httptest.NewRequest("POST", "/api/proposals/7/votes", strings.NewReader(good))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("X-User-Id", "1")
	rr2 := httptest.NewRecorder()
	r.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusCreated {
		t.Fatalf("create: want 201 got %d", rr2.Code)
	}

	// Duplicate
	req3 := httptest.NewRequest("POST", "/api/proposals/7/votes", strings.NewReader(good))
	req3.Header.Set("Content-Type", "application/json")
	req3.Header.Set("X-User-Id", "1")
	rr3 := httptest.NewRecorder()
	r.ServeHTTP(rr3, req3)
	if rr3.Code != http.StatusConflict {
		t.Fatalf("duplicate: want 409 got %d", rr3.Code)
	}
}

func TestVotes_TallyQuorum(t *testing.T) {
	repo := &mockRepo{}
	r := testRouter(repo)

	// Cast 3 votes (below quorum of 6)
	for i, choice := range []string{"for", "against", "for"} {
		body := bytes.NewBufferString(`{"choice":"` + choice + `"}`)
		req := httptest.NewRequest("POST", "/api/proposals/9/votes", body)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-Id", strconv.Itoa(i+1))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		if rr.Code != http.StatusCreated {
			t.Fatalf("seed vote %d: want 201 got %d", i, rr.Code)
		}
	}

	// Tally should reflect counts and quorum=false
	reqT := httptest.NewRequest("GET", "/api/proposals/9/votes/tally", nil)
	rrT := httptest.NewRecorder()
	r.ServeHTTP(rrT, reqT)
	if rrT.Code != http.StatusOK {
		t.Fatalf("tally: want 200 got %d", rrT.Code)
	}
	var tally Tally
	_ = json.Unmarshal(rrT.Body.Bytes(), &tally)
	if tally.Results["for"] != 2 || tally.Results["against"] != 1 || tally.QuorumMet {
		t.Fatalf("unexpected tally: %+v", tally)
	}
}
