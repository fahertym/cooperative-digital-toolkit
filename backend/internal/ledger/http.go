package ledger

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	Repo Repo
}

func (h Handlers) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.Repo.List(r.Context())
	if err != nil {
		http.Error(w, "failed to list", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(items)
}

func (h Handlers) Create(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Type        string  `json:"type"`
		Amount      float64 `json:"amount"`
		Description string  `json:"description"`
		MemberID    *int32  `json:"member_id"`
		Notes       string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if len(in.Type) == 0 || len(in.Description) == 0 {
		http.Error(w, "type and description required", http.StatusBadRequest)
		return
	}
	e, err := h.Repo.Create(r.Context(), in.Type, in.Amount, in.Description, in.MemberID, in.Notes)
	if err != nil {
		http.Error(w, "insert failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(e)
}

func (h Handlers) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id64, _ := strconv.ParseInt(idStr, 10, 32)
	e, err := h.Repo.Get(r.Context(), int32(id64))
	if err != nil {
		if err == ErrNotFound {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "query failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(e)
}

// ExportCSV streams all ledger entries as CSV
func (h Handlers) ExportCSV(w http.ResponseWriter, r *http.Request) {
	items, err := h.Repo.List(r.Context())
	if err != nil {
		http.Error(w, "failed to list", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=ledger.csv")

	cw := csv.NewWriter(w)
	defer cw.Flush()

	_ = cw.Write([]string{"id", "type", "amount", "description", "member_id", "notes", "created_at"})
	for _, e := range items {
		memberID := ""
		if e.MemberID != nil {
			memberID = strconv.FormatInt(int64(*e.MemberID), 10)
		}
		_ = cw.Write([]string{
			strconv.FormatInt(int64(e.ID), 10),
			e.Type,
			strconv.FormatFloat(e.Amount, 'f', 2, 64),
			e.Description,
			memberID,
			e.Notes,
			e.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
}
