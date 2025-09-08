package announcements

import (
    "encoding/json"
    "net/http"
    "strconv"

    "coop.tools/backend/internal/httpmw"
    "coop.tools/backend/internal/httpx"
    "github.com/go-chi/chi/v5"
)

type Handlers struct {
	Repo Repo
}

func (h Handlers) List(w http.ResponseWriter, r *http.Request) {
    var memberID *int64
    if mid := r.URL.Query().Get("member_id"); mid != "" {
        if v, err := strconv.ParseInt(mid, 10, 64); err == nil {
            memberID = &v
        } else {
            httpmw.WriteJSONError(w, http.StatusBadRequest, "invalid member_id")
            return
        }
    }
    filters := &ListFilters{}
    if p := httpx.QueryString(r, "priority"); p != "" { filters.Priority = p }
    if v, err := httpx.QueryInt64(r, "author_id"); err != nil {
        httpmw.WriteJSONError(w, http.StatusBadRequest, "invalid author_id")
        return
    } else if v != nil { filters.AuthorID = v }
    if httpx.QueryBoolTrue(r, "only_unread") { filters.OnlyUnread = true }

	// Optional pagination
    if lim, off, err := httpx.ParseLimitOffset(r, 200); err != nil {
        httpmw.WriteJSONError(w, http.StatusBadRequest, "invalid pagination")
        return
    } else { filters.Limit, filters.Offset = lim, off }

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

    // Author is the authenticated admin user
    p, _ := httpmw.FromContext(r.Context())
    authorID := p.MemberID
    a, err := h.Repo.Create(r.Context(), in.Title, in.Body, &authorID, in.Priority)
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
    var memberID *int64
    if mid := r.URL.Query().Get("member_id"); mid != "" {
        if v, err := strconv.ParseInt(mid, 10, 64); err == nil {
            memberID = &v
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
    p, ok := httpmw.FromContext(r.Context())
    if !ok || p.MemberID <= 0 {
        httpmw.WriteJSONError(w, http.StatusUnauthorized, "unauthorized")
        return
    }
    if err := h.Repo.MarkAsRead(r.Context(), int32(id64), p.MemberID); err != nil {
        if err == ErrNotFound {
            httpmw.WriteJSONError(w, http.StatusNotFound, "not found")
            return
        }
        httpmw.WriteJSONError(w, http.StatusInternalServerError, "failed to mark read")
        return
    }
    m := p.MemberID
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
    v, err := strconv.ParseInt(mid, 10, 64)
    if err != nil {
        httpmw.WriteJSONError(w, http.StatusBadRequest, "invalid member_id")
        return
    }
    memberID := int64(v)
    count, err := h.Repo.GetUnreadCount(r.Context(), memberID)
    if err != nil {
        httpmw.WriteJSONError(w, http.StatusInternalServerError, "failed to get unread count")
        return
    }
	w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(struct {
        MemberID    int64 `json:"member_id"`
        UnreadCount int   `json:"unread_count"`
    }{MemberID: memberID, UnreadCount: count})
}
