package members

import (
    "encoding/json"
    "net/http"
    "strconv"

    "coop.tools/backend/internal/httpmw"
    "github.com/go-chi/chi/v5"
)

type Handlers struct{ Repo Repo }

// GetByID handles GET /api/members/{id}
func (h Handlers) GetByID(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil || id <= 0 {
        httpmw.WriteJSONError(w, http.StatusBadRequest, "invalid id")
        return
    }
    m, err := h.Repo.GetByID(r.Context(), id)
    if err != nil {
        if err == ErrNotFound {
            httpmw.WriteJSONError(w, http.StatusNotFound, "not found")
            return
        }
        httpmw.WriteJSONError(w, http.StatusInternalServerError, "query failed")
        return
    }
    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(m)
}

// FindByEmail handles GET /api/members?email=
func (h Handlers) FindByEmail(w http.ResponseWriter, r *http.Request) {
    email := r.URL.Query().Get("email")
    if email == "" {
        httpmw.WriteJSONError(w, http.StatusBadRequest, "email required")
        return
    }
    m, err := h.Repo.GetByEmail(r.Context(), email)
    if err != nil {
        if err == ErrNotFound {
            httpmw.WriteJSONError(w, http.StatusNotFound, "not found")
            return
        }
        httpmw.WriteJSONError(w, http.StatusInternalServerError, "query failed")
        return
    }
    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(m)
}

// Create handles POST /api/members
func (h Handlers) Create(w http.ResponseWriter, r *http.Request) {
    var in struct {
        Email       string `json:"email"`
        DisplayName string `json:"display_name"`
        Role        string `json:"role"`
    }
    if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
        httpmw.WriteJSONError(w, http.StatusBadRequest, "invalid json")
        return
    }
    if in.Email == "" || in.DisplayName == "" {
        httpmw.WriteJSONError(w, http.StatusBadRequest, "email and display_name required")
        return
    }
    role := in.Role
    if role == "" { role = "member" }
    if role != "admin" && role != "member" {
        httpmw.WriteJSONError(w, http.StatusBadRequest, "invalid role")
        return
    }
    m, err := h.Repo.Create(r.Context(), in.Email, in.DisplayName, role)
    if err != nil {
        if err == ErrConflict {
            httpmw.WriteJSONError(w, http.StatusConflict, "email already exists")
            return
        }
        httpmw.WriteJSONError(w, http.StatusInternalServerError, "create failed")
        return
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    _ = json.NewEncoder(w).Encode(m)
}

