package proposals

import (
    "encoding/csv"
    "encoding/json"
    "errors"
    "net/http"
    "strconv"
    "time"

    "coop.tools/backend/internal/httpmw"
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
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if len(in.Title) == 0 {
		http.Error(w, "title required", http.StatusBadRequest)
		return
	}
	p, err := h.Repo.Create(r.Context(), in.Title, in.Body)
	if err != nil {
		http.Error(w, "insert failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(p)
}

func (h Handlers) Get(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id64, err := strconv.ParseInt(idStr, 10, 32)
    if err != nil {
        httpmw.WriteJSONError(w, http.StatusBadRequest, "invalid id")
        return
    }
    p, err := h.Repo.Get(r.Context(), int32(id64))
	if err != nil {
		if err == ErrNotFound {
            httpmw.WriteJSONError(w, http.StatusNotFound, "not found")
            return
        }
        httpmw.WriteJSONError(w, http.StatusInternalServerError, "query failed")
        return
    }
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(p)
}

// Close transitions a proposal from open to closed.
// POST /api/proposals/{id}/close
func (h Handlers) Close(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id64, err := strconv.ParseInt(idStr, 10, 32)
    if err != nil {
        httpmw.WriteJSONError(w, http.StatusBadRequest, "invalid id")
        return
    }
    p, err := h.Repo.Close(r.Context(), int32(id64))
	switch {
	case err == nil:
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(p)
        case errors.Is(err, ErrNotFound):
            httpmw.WriteJSONError(w, http.StatusNotFound, "not found")
        case errors.Is(err, ErrConflict):
            httpmw.WriteJSONError(w, http.StatusConflict, "proposal not open")
        default:
            httpmw.WriteJSONError(w, http.StatusInternalServerError, "close failed")
        }
}

// ExportCSV streams all proposals as CSV
func (h Handlers) ExportCSV(w http.ResponseWriter, r *http.Request) {
	items, err := h.Repo.List(r.Context())
	if err != nil {
		http.Error(w, "failed to list", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=proposals.csv")

	cw := csv.NewWriter(w)
	defer cw.Flush()

	_ = cw.Write([]string{"id", "title", "body", "status", "created_at"})
	for _, p := range items {
		_ = cw.Write([]string{
			strconv.FormatInt(int64(p.ID), 10),
			p.Title,
			p.Body,
			p.Status,
			p.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
}
