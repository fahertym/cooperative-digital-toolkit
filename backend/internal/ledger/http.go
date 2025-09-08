package ledger

import (
    "encoding/csv"
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
    filters := &ListFilters{}
    if t := r.URL.Query().Get("type"); t != "" {
        filters.Type = t
    }
    if mid := r.URL.Query().Get("member_id"); mid != "" {
        v, err := strconv.ParseInt(mid, 10, 32)
        if err != nil {
            httpmw.WriteJSONError(w, http.StatusBadRequest, "invalid member_id")
            return
        }
        v32 := int32(v)
        filters.MemberID = &v32
    }
    if ls := r.URL.Query().Get("limit"); ls != "" {
        if v, err := strconv.Atoi(ls); err == nil && v > 0 {
            if v > 200 {
                v = 200
            }
            filters.Limit = v
        } else if err != nil {
            httpmw.WriteJSONError(w, http.StatusBadRequest, "invalid limit")
            return
        }
    }
    if os := r.URL.Query().Get("offset"); os != "" {
        if v, err := strconv.Atoi(os); err == nil && v >= 0 {
            filters.Offset = v
        } else if err != nil {
            httpmw.WriteJSONError(w, http.StatusBadRequest, "invalid offset")
            return
        }
    }

    items, err := h.Repo.List(r.Context(), filters)
    if err != nil {
        httpmw.WriteJSONError(w, http.StatusInternalServerError, "failed to list")
        return
    }
    if filters.Limit > 0 { w.Header().Set("X-Limit", strconv.Itoa(filters.Limit)) }
    if filters.Offset > 0 { w.Header().Set("X-Offset", strconv.Itoa(filters.Offset)) }
    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(items)
}

func (h Handlers) Create(w http.ResponseWriter, r *http.Request) {
	memberID, ok := httpmw.CurrentUserID(r.Context())
    if !ok {
        httpmw.WriteJSONError(w, http.StatusUnauthorized, "unauthorized")
        return
    }
	var in struct {
		Type        string  `json:"type"`
		Amount      float64 `json:"amount"`
		Description string  `json:"description"`
		Notes       string  `json:"notes"`
	}
    if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
        httpmw.WriteJSONError(w, http.StatusBadRequest, "invalid json")
        return
    }
    if len(in.Type) == 0 || len(in.Description) == 0 {
        httpmw.WriteJSONError(w, http.StatusBadRequest, "type and description required")
        return
    }
    if in.Amount == 0 {
        httpmw.WriteJSONError(w, http.StatusBadRequest, "amount must be non-zero")
        return
    }
    if in.Type != "dues" && in.Type != "contribution" && in.Type != "expense" && in.Type != "income" {
        httpmw.WriteJSONError(w, http.StatusBadRequest, "invalid type")
        return
    }
	idem := r.Header.Get("X-Idempotency-Key")
	mid := memberID
    e, replayed, err := h.Repo.Create(r.Context(), in.Type, in.Description, in.Amount, &mid, in.Notes, idem)
    if err != nil {
        httpmw.WriteJSONError(w, http.StatusInternalServerError, "insert failed")
        return
    }
    w.Header().Set("Content-Type", "application/json")
    if replayed {
        // Idempotent replay: return 200 to indicate existing resource
        w.WriteHeader(http.StatusOK)
    } else {
        w.WriteHeader(http.StatusCreated)
    }
    _ = json.NewEncoder(w).Encode(e)
}

func (h Handlers) Get(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id64, err := strconv.ParseInt(idStr, 10, 32)
    if err != nil {
        httpmw.WriteJSONError(w, http.StatusBadRequest, "invalid id")
        return
    }
	e, err := h.Repo.Get(r.Context(), int32(id64))
    if err != nil {
        if err == ErrNotFound {
            httpmw.WriteJSONError(w, http.StatusNotFound, "not found")
            return
        }
        httpmw.WriteJSONError(w, http.StatusInternalServerError, "query failed")
        return
    }
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(e)
}

// ExportCSV streams all ledger entries as CSV
func (h Handlers) ExportCSV(w http.ResponseWriter, r *http.Request) {
	items, err := h.Repo.List(r.Context(), &ListFilters{})
	if err != nil {
		http.Error(w, "failed to list", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=ledger.csv")

	cw := csv.NewWriter(w)
	defer cw.Flush()

	_ = cw.Write([]string{"Date", "Description", "Type", "Amount", "Member ID", "Notes", "Reference"})
	for _, e := range items {
		memberID := ""
		if e.MemberID != nil {
			memberID = strconv.FormatInt(int64(*e.MemberID), 10)
		}
		_ = cw.Write([]string{
			e.CreatedAt.UTC().Format("2006-01-02"),
			e.Description,
			e.Type,
			strconv.FormatFloat(e.Amount, 'f', 2, 64),
			memberID,
			e.Notes,
			strconv.FormatInt(int64(e.ID), 10),
		})
	}
}
