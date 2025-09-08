package announcements

import (
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

	if priority := r.URL.Query().Get("priority"); priority != "" {
		filters.Priority = priority
	}

	if authorIDStr := r.URL.Query().Get("author_id"); authorIDStr != "" {
		if authorID, err := strconv.ParseInt(authorIDStr, 10, 32); err == nil {
			id := int32(authorID)
			filters.AuthorID = &id
		}
	}

	if r.URL.Query().Get("unread") == "true" {
		filters.OnlyUnread = true
	}

	if fromDate := r.URL.Query().Get("from_date"); fromDate != "" {
		filters.FromDate = &fromDate
	}

	if toDate := r.URL.Query().Get("to_date"); toDate != "" {
		filters.ToDate = &toDate
	}

	// For now, use a mock member ID. This will be replaced with authenticated member ID
	// when authentication is implemented
	var memberID *int32
	if memberIDStr := r.URL.Query().Get("member_id"); memberIDStr != "" {
		if parsedID, err := strconv.ParseInt(memberIDStr, 10, 32); err == nil {
			id := int32(parsedID)
			memberID = &id
		}
	}

	announcements, err := h.Repo.List(r.Context(), memberID, filters)
	if err != nil {
		http.Error(w, "failed to list announcements", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(announcements)
}

func (h Handlers) Create(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title    string `json:"title"`
		Body     string `json:"body"`
		AuthorID *int32 `json:"author_id"`
		Priority string `json:"priority"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if in.Title == "" {
		http.Error(w, "title required", http.StatusBadRequest)
		return
	}
	if in.Body == "" {
		http.Error(w, "body required", http.StatusBadRequest)
		return
	}

	// Validate priority enum, default to 'normal'
	if in.Priority == "" {
		in.Priority = "normal"
	}
	validPriorities := map[string]bool{"low": true, "normal": true, "high": true, "urgent": true}
	if !validPriorities[in.Priority] {
		http.Error(w, "invalid priority, must be one of: low, normal, high, urgent", http.StatusBadRequest)
		return
	}

	announcement, err := h.Repo.Create(r.Context(), in.Title, in.Body, in.AuthorID, in.Priority)
	if err != nil {
		http.Error(w, "insert failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(announcement)
}

func (h Handlers) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id64, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	// For now, use query parameter for member ID. This will be replaced with authenticated member ID
	// when authentication is implemented
	var memberID *int32
	if memberIDStr := r.URL.Query().Get("member_id"); memberIDStr != "" {
		if parsedID, err := strconv.ParseInt(memberIDStr, 10, 32); err == nil {
			id := int32(parsedID)
			memberID = &id
		}
	}

	announcement, err := h.Repo.Get(r.Context(), int32(id64), memberID)
	if err != nil {
		if err == ErrNotFound {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "query failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(announcement)
}

func (h Handlers) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	announcementID64, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		http.Error(w, "invalid announcement id", http.StatusBadRequest)
		return
	}

	var in struct {
		MemberID int32 `json:"member_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if in.MemberID == 0 {
		http.Error(w, "member_id required", http.StatusBadRequest)
		return
	}

	// First verify the announcement exists
	_, err = h.Repo.Get(r.Context(), int32(announcementID64), &in.MemberID)
	if err != nil {
		if err == ErrNotFound {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "failed to verify announcement", http.StatusInternalServerError)
		return
	}

	// Mark as read
	if err := h.Repo.MarkAsRead(r.Context(), int32(announcementID64), in.MemberID); err != nil {
		http.Error(w, "failed to mark as read", http.StatusInternalServerError)
		return
	}

	// Return the updated announcement with read status
	updatedAnnouncement, err := h.Repo.Get(r.Context(), int32(announcementID64), &in.MemberID)
	if err != nil {
		http.Error(w, "failed to retrieve updated announcement", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(updatedAnnouncement)
}

func (h Handlers) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	memberIDStr := r.URL.Query().Get("member_id")
	if memberIDStr == "" {
		http.Error(w, "member_id query parameter required", http.StatusBadRequest)
		return
	}

	memberID64, err := strconv.ParseInt(memberIDStr, 10, 32)
	if err != nil {
		http.Error(w, "invalid member_id", http.StatusBadRequest)
		return
	}

	count, err := h.Repo.GetUnreadCount(r.Context(), int32(memberID64))
	if err != nil {
		http.Error(w, "failed to get unread count", http.StatusInternalServerError)
		return
	}

	response := struct {
		MemberID    int32 `json:"member_id"`
		UnreadCount int   `json:"unread_count"`
	}{
		MemberID:    int32(memberID64),
		UnreadCount: count,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}
