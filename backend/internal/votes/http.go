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
    proposalID64, err := strconv.ParseInt(proposalIDStr, 10, 32)
    if err != nil {
        httpmw.WriteJSONError(w, http.StatusBadRequest, "invalid proposal_id")
        return
    }

    var limit, offset int
    if ls := r.URL.Query().Get("limit"); ls != "" {
        if v, err := strconv.Atoi(ls); err == nil && v > 0 {
            if v > 200 { v = 200 }
            limit = v
        } else if err != nil {
            httpmw.WriteJSONError(w, http.StatusBadRequest, "invalid limit")
            return
        }
    }
    if os := r.URL.Query().Get("offset"); os != "" {
        if v, err := strconv.Atoi(os); err == nil && v >= 0 {
            offset = v
        } else if err != nil {
            httpmw.WriteJSONError(w, http.StatusBadRequest, "invalid offset")
            return
        }
    }

    items, err := h.Repo.List(r.Context(), int32(proposalID64), limit, offset)
    if err != nil {
        httpmw.WriteJSONError(w, http.StatusInternalServerError, "failed to list votes")
        return
    }
    if limit > 0 { w.Header().Set("X-Limit", strconv.Itoa(limit)) }
    if offset > 0 { w.Header().Set("X-Offset", strconv.Itoa(offset)) }
    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(items)
}

func (h Handlers) Create(w http.ResponseWriter, r *http.Request) {
    proposalIDStr := chi.URLParam(r, "proposal_id")
    proposalID64, err := strconv.ParseInt(proposalIDStr, 10, 32)
    if err != nil {
        http.Error(w, "invalid proposal_id", http.StatusBadRequest)
        return
    }

	var in struct {
		Choice string `json:"choice"`
		Notes  string `json:"notes"`
	}
    if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
        httpmw.WriteJSONError(w, http.StatusBadRequest, "invalid json")
        return
    }

	// Validate choice
    if in.Choice != "for" && in.Choice != "against" && in.Choice != "abstain" {
        httpmw.WriteJSONError(w, http.StatusBadRequest, "choice must be 'for', 'against', or 'abstain'")
        return
    }

	uID, ok := httpmw.CurrentUserID(r.Context())
    if !ok {
        httpmw.WriteJSONError(w, http.StatusUnauthorized, "unauthorized")
        return
    }
	v, err := h.Repo.Create(r.Context(), int32(proposalID64), uID, in.Choice, in.Notes)
	if err != nil {
        switch err {
        case ErrNotFound:
            httpmw.WriteJSONError(w, http.StatusNotFound, "not found")
        case ErrAlreadyVoted:
            httpmw.WriteJSONError(w, http.StatusConflict, "member already voted on this proposal")
        case ErrProposalClosed:
            httpmw.WriteJSONError(w, http.StatusConflict, "proposal is closed")
        default:
            httpmw.WriteJSONError(w, http.StatusInternalServerError, "failed to create vote")
        }
        return
    }

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(v)
}

func (h Handlers) Update(w http.ResponseWriter, r *http.Request) {
    proposalIDStr := chi.URLParam(r, "proposal_id")
    proposalID64, err := strconv.ParseInt(proposalIDStr, 10, 32)
    if err != nil {
        http.Error(w, "invalid proposal_id", http.StatusBadRequest)
        return
    }

	var in struct {
		Choice string `json:"choice"`
		Notes  string `json:"notes"`
	}
    if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
        httpmw.WriteJSONError(w, http.StatusBadRequest, "invalid json")
        return
    }

	// Validate choice
    if in.Choice != "for" && in.Choice != "against" && in.Choice != "abstain" {
        httpmw.WriteJSONError(w, http.StatusBadRequest, "choice must be 'for', 'against', or 'abstain'")
        return
    }

	uID, ok := httpmw.CurrentUserID(r.Context())
    if !ok {
        httpmw.WriteJSONError(w, http.StatusUnauthorized, "unauthorized")
        return
    }
	v, err := h.Repo.Update(r.Context(), int32(proposalID64), uID, in.Choice, in.Notes)
	if err != nil {
        switch err {
        case ErrNotFound:
            httpmw.WriteJSONError(w, http.StatusNotFound, "not found")
        case ErrProposalClosed:
            httpmw.WriteJSONError(w, http.StatusConflict, "proposal is closed")
        default:
            httpmw.WriteJSONError(w, http.StatusInternalServerError, "failed to update vote")
        }
        return
    }

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func (h Handlers) GetTally(w http.ResponseWriter, r *http.Request) {
    proposalIDStr := chi.URLParam(r, "proposal_id")
    proposalID64, err := strconv.ParseInt(proposalIDStr, 10, 32)
    if err != nil {
        http.Error(w, "invalid proposal_id", http.StatusBadRequest)
        return
    }

	tally, err := h.Repo.GetTally(r.Context(), int32(proposalID64))
    if err != nil {
        if err == ErrNotFound {
            httpmw.WriteJSONError(w, http.StatusNotFound, "not found")
            return
        }
        httpmw.WriteJSONError(w, http.StatusInternalServerError, "failed to get tally")
        return
    }

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tally)
}
