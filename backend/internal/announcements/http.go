package announcements

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

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
		Title    string `json:"title"`
		Body     string `json:"body"`
		AuthorID int32  `json:"author_id"`
		Priority string `json:"priority"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if len(in.Title) == 0 || len(in.Body) == 0 {
		http.Error(w, "title and body required", http.StatusBadRequest)
		return
	}
	if in.Priority == "" {
		in.Priority = "normal"
	}
	a, err := h.Repo.Create(r.Context(), in.Title, in.Body, in.AuthorID, in.Priority)
	if err != nil {
		http.Error(w, "insert failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(a)
}

func (h Handlers) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id64, _ := strconv.ParseInt(idStr, 10, 32)
	a, err := h.Repo.Get(r.Context(), int32(id64))
	if err != nil {
		if err == ErrNotFound {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "query failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(a)
}