package announcements

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
	var memberID *int32
	if mid := r.URL.Query().Get("member_id"); mid != "" {
		if v, err := strconv.ParseInt(mid, 10, 32); err == nil {
			v32 := int32(v)
			memberID = &v32
        } else {
            httpmw.WriteJSONError(w, http.StatusBadRequest, "invalid member_id")
            return
        }
	}
	filters := &ListFilters{}
	if p := r.URL.Query().Get("priority"); p != "" {
		filters.Priority = p
	}
	if aid := r.URL.Query().Get("author_id"); aid != "" {
		v, err := strconv.ParseInt(aid, 10, 32)
		if err != nil {
			http.Error(w, "invalid author_id", http.StatusBadRequest)
			return
		}
		v32 := int32(v)
		filters.AuthorID = &v32
	}
	if r.URL.Query().Get("only_unread") == "true" {
		filters.OnlyUnread = true
	}

	// Optional pagination
	if ls := r.URL.Query().Get("limit"); ls != "" {
		if v, err := strconv.Atoi(ls); err == nil && v > 0 {
			if v > 200 { v = 200 }
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

    items, err := h.Repo.List(r.Context(), memberID, filters)
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
	var in struct {
		Title    string `json:"title"`
		Body     string `json:"body"`
		AuthorID *int32 `json:"author_id"`
		Priority string `json:"priority"`
	}
    if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
        httpmw.WriteJSONError(w, http.StatusBadRequest, "invalid json")
        return
    }
    if len(in.Title) == 0 || len(in.Body) == 0 {
        httpmw.WriteJSONError(w, http.StatusBadRequest, "title and body required")
        return
    }
	if in.Priority == "" {
		in.Priority = "normal"
	}
    if in.Priority != "low" && in.Priority != "normal" && in.Priority != "high" && in.Priority != "urgent" {
        httpmw.WriteJSONError(w, http.StatusBadRequest, "invalid priority")
        return
    }

	a, err := h.Repo.Create(r.Context(), in.Title, in.Body, in.AuthorID, in.Priority)
    if err != nil {
        httpmw.WriteJSONError(w, http.StatusInternalServerError, "insert failed")
        return
    }
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(a)
}

func (h Handlers) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id64, err := strconv.ParseInt(idStr, 10, 32)
    if err != nil {
        httpmw.WriteJSONError(w, http.StatusBadRequest, "invalid id")
        return
    }
	var memberID *int32
	if mid := r.URL.Query().Get("member_id"); mid != "" {
		if v, err := strconv.ParseInt(mid, 10, 32); err == nil {
			v32 := int32(v)
			memberID = &v32
        } else {
            httpmw.WriteJSONError(w, http.StatusBadRequest, "invalid member_id")
            return
        }
	}
	a, err := h.Repo.Get(r.Context(), int32(id64), memberID)
    if err != nil {
        if err == ErrNotFound {
            httpmw.WriteJSONError(w, http.StatusNotFound, "not found")
            return
        }
        httpmw.WriteJSONError(w, http.StatusInternalServerError, "query failed")
        return
    }
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(a)
}

// MarkRead marks an announcement as read for a member.
func (h Handlers) MarkRead(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id64, err := strconv.ParseInt(idStr, 10, 32)
    if err != nil {
        httpmw.WriteJSONError(w, http.StatusBadRequest, "invalid id")
        return
    }
	uid, ok := httpmw.CurrentUserID(r.Context())
    if !ok {
        httpmw.WriteJSONError(w, http.StatusUnauthorized, "unauthorized")
        return
    }
    if err := h.Repo.MarkAsRead(r.Context(), int32(id64), uid); err != nil {
        if err == ErrNotFound {
            httpmw.WriteJSONError(w, http.StatusNotFound, "not found")
            return
        }
        httpmw.WriteJSONError(w, http.StatusInternalServerError, "failed to mark read")
        return
    }
	m := uid
	a, err := h.Repo.Get(r.Context(), int32(id64), &m)
    if err != nil {
        if err == ErrNotFound {
            httpmw.WriteJSONError(w, http.StatusNotFound, "not found")
            return
        }
        httpmw.WriteJSONError(w, http.StatusInternalServerError, "query failed")
        return
    }
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(a)
}

// UnreadCount returns unread count for a member.
func (h Handlers) UnreadCount(w http.ResponseWriter, r *http.Request) {
	mid := r.URL.Query().Get("member_id")
    if mid == "" {
        httpmw.WriteJSONError(w, http.StatusBadRequest, "member_id required")
        return
    }
	v, err := strconv.ParseInt(mid, 10, 32)
    if err != nil {
        httpmw.WriteJSONError(w, http.StatusBadRequest, "invalid member_id")
        return
    }
	memberID := int32(v)
	count, err := h.Repo.GetUnreadCount(r.Context(), memberID)
    if err != nil {
        httpmw.WriteJSONError(w, http.StatusInternalServerError, "failed to get unread count")
        return
    }
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		MemberID    int32 `json:"member_id"`
		UnreadCount int   `json:"unread_count"`
	}{MemberID: memberID, UnreadCount: count})
}
