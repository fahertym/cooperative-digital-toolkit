package votes

import (
	"encoding/json"
	"net/http"
	"strconv"

	"coop.tools/backend/internal/httpmw"
	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	Repo Repo
}

func (h Handlers) List(w http.ResponseWriter, r *http.Request) {
	proposalIDStr := chi.URLParam(r, "proposal_id")
	proposalID64, _ := strconv.ParseInt(proposalIDStr, 10, 32)

	items, err := h.Repo.List(r.Context(), int32(proposalID64))
	if err != nil {
		http.Error(w, "failed to list votes", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(items)
}

func (h Handlers) Create(w http.ResponseWriter, r *http.Request) {
	proposalIDStr := chi.URLParam(r, "proposal_id")
	proposalID64, _ := strconv.ParseInt(proposalIDStr, 10, 32)

	var in struct {
		Choice string `json:"choice"`
		Notes  string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// Validate choice
	if in.Choice != "for" && in.Choice != "against" && in.Choice != "abstain" {
		http.Error(w, "choice must be 'for', 'against', or 'abstain'", http.StatusBadRequest)
		return
	}

	uID, ok := httpmw.CurrentUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	v, err := h.Repo.Create(r.Context(), int32(proposalID64), uID, in.Choice, in.Notes)
	if err != nil {
		switch err {
		case ErrNotFound:
			http.NotFound(w, r)
		case ErrAlreadyVoted:
			http.Error(w, "member already voted on this proposal", http.StatusConflict)
		case ErrProposalClosed:
			http.Error(w, "proposal is closed", http.StatusConflict)
		default:
			http.Error(w, "failed to create vote", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(v)
}

func (h Handlers) Update(w http.ResponseWriter, r *http.Request) {
	proposalIDStr := chi.URLParam(r, "proposal_id")
	proposalID64, _ := strconv.ParseInt(proposalIDStr, 10, 32)

	var in struct {
		Choice string `json:"choice"`
		Notes  string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// Validate choice
	if in.Choice != "for" && in.Choice != "against" && in.Choice != "abstain" {
		http.Error(w, "choice must be 'for', 'against', or 'abstain'", http.StatusBadRequest)
		return
	}

	uID, ok := httpmw.CurrentUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	v, err := h.Repo.Update(r.Context(), int32(proposalID64), uID, in.Choice, in.Notes)
	if err != nil {
		switch err {
		case ErrNotFound:
			http.NotFound(w, r)
		case ErrProposalClosed:
			http.Error(w, "proposal is closed", http.StatusConflict)
		default:
			http.Error(w, "failed to update vote", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func (h Handlers) GetTally(w http.ResponseWriter, r *http.Request) {
	proposalIDStr := chi.URLParam(r, "proposal_id")
	proposalID64, _ := strconv.ParseInt(proposalIDStr, 10, 32)

	tally, err := h.Repo.GetTally(r.Context(), int32(proposalID64))
	if err != nil {
		if err == ErrNotFound {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "failed to get tally", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tally)
}
