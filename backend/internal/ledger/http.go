package ledger

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	Repo Repo
}

func (h Handlers) List(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters for filtering
	filters := &ListFilters{}

	if entryType := r.URL.Query().Get("type"); entryType != "" {
		filters.Type = entryType
	}

	if memberIDStr := r.URL.Query().Get("member_id"); memberIDStr != "" {
		if memberID, err := strconv.ParseInt(memberIDStr, 10, 32); err == nil {
			id := int32(memberID)
			filters.MemberID = &id
		}
	}

	if fromDate := r.URL.Query().Get("from_date"); fromDate != "" {
		filters.FromDate = &fromDate
	}

	if toDate := r.URL.Query().Get("to_date"); toDate != "" {
		filters.ToDate = &toDate
	}

	entries, err := h.Repo.List(r.Context(), filters)
	if err != nil {
		http.Error(w, "failed to list entries", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(entries)
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

	// Validate required fields
	if in.Type == "" {
		http.Error(w, "type required", http.StatusBadRequest)
		return
	}
	if in.Description == "" {
		http.Error(w, "description required", http.StatusBadRequest)
		return
	}
	if in.Amount == 0 {
		http.Error(w, "amount required and cannot be zero", http.StatusBadRequest)
		return
	}

	// Validate type enum
	validTypes := map[string]bool{"dues": true, "contribution": true, "expense": true, "income": true}
	if !validTypes[in.Type] {
		http.Error(w, "invalid type, must be one of: dues, contribution, expense, income", http.StatusBadRequest)
		return
	}

	entry, err := h.Repo.Create(r.Context(), in.Type, in.Description, in.Amount, in.MemberID, in.Notes)
	if err != nil {
		http.Error(w, "insert failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(entry)
}

func (h Handlers) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id64, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	entry, err := h.Repo.Get(r.Context(), int32(id64))
	if err != nil {
		if err == ErrNotFound {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "query failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(entry)
}

// ExportCSV streams all ledger entries as CSV compatible with QuickBooks/Xero
func (h Handlers) ExportCSV(w http.ResponseWriter, r *http.Request) {
	// Parse same filters as List for consistent filtering
	filters := &ListFilters{}

	if entryType := r.URL.Query().Get("type"); entryType != "" {
		filters.Type = entryType
	}

	if memberIDStr := r.URL.Query().Get("member_id"); memberIDStr != "" {
		if memberID, err := strconv.ParseInt(memberIDStr, 10, 32); err == nil {
			id := int32(memberID)
			filters.MemberID = &id
		}
	}

	if fromDate := r.URL.Query().Get("from_date"); fromDate != "" {
		filters.FromDate = &fromDate
	}

	if toDate := r.URL.Query().Get("to_date"); toDate != "" {
		filters.ToDate = &toDate
	}

	entries, err := h.Repo.List(r.Context(), filters)
	if err != nil {
		http.Error(w, "failed to list entries", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=ledger.csv")

	cw := csv.NewWriter(w)
	defer cw.Flush()

	// QuickBooks/Xero compatible headers
	_ = cw.Write([]string{"Date", "Description", "Type", "Amount", "Member ID", "Notes", "Reference"})

	for _, entry := range entries {
		memberIDStr := ""
		if entry.MemberID != nil {
			memberIDStr = strconv.FormatInt(int64(*entry.MemberID), 10)
		}

		_ = cw.Write([]string{
			entry.CreatedAt.UTC().Format("2006-01-02"), // Date format for accounting software
			entry.Description,
			entry.Type,
			strconv.FormatFloat(entry.Amount, 'f', 2, 64), // 2 decimal places
			memberIDStr,
			entry.Notes,
			strconv.FormatInt(int64(entry.ID), 10), // Reference/ID
		})
	}
}
