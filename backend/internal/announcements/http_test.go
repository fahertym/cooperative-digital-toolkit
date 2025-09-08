package announcements

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

// ---- Mock Repo ----

type mockRepo struct {
	announcements []Announcement
	reads         []AnnouncementRead
	nextID        int32
}

func (m *mockRepo) List(ctx context.Context, memberID *int32, filters *ListFilters) ([]AnnouncementWithReadStatus, error) {
	var out []AnnouncementWithReadStatus

	for _, announcement := range m.announcements {
		include := true

		// Apply filters
		if filters != nil {
			if filters.Priority != "" && announcement.Priority != filters.Priority {
				include = false
			}
			if filters.AuthorID != nil {
				if announcement.AuthorID == nil || *announcement.AuthorID != *filters.AuthorID {
					include = false
				}
			}
			// Skip date filters in mock for simplicity
		}

		if include {
			item := AnnouncementWithReadStatus{
				Announcement: announcement,
				IsRead:       false,
				ReadAt:       nil,
			}
			if memberID != nil {
				item.MemberID = *memberID
				// Check if this member has read this announcement
				for _, read := range m.reads {
					if read.AnnouncementID == announcement.ID && read.MemberID == *memberID {
						item.IsRead = true
						item.ReadAt = &read.ReadAt
						break
					}
				}
			}
			
			// Apply unread filter
			if filters != nil && filters.OnlyUnread && item.IsRead {
				continue
			}
			
			out = append(out, item)
		}
	}

	return out, nil
}

func (m *mockRepo) Get(ctx context.Context, id int32, memberID *int32) (AnnouncementWithReadStatus, error) {
	for _, announcement := range m.announcements {
		if announcement.ID == id {
			item := AnnouncementWithReadStatus{
				Announcement: announcement,
				IsRead:       false,
				ReadAt:       nil,
			}
			if memberID != nil {
				item.MemberID = *memberID
				// Check if this member has read this announcement
				for _, read := range m.reads {
					if read.AnnouncementID == id && read.MemberID == *memberID {
						item.IsRead = true
						item.ReadAt = &read.ReadAt
						break
					}
				}
			}
			return item, nil
		}
	}
	return AnnouncementWithReadStatus{}, ErrNotFound
}

func (m *mockRepo) Create(ctx context.Context, title, body string, authorID *int32, priority string) (Announcement, error) {
	if m.nextID == 0 {
		m.nextID = 1
	}
	
	now := time.Now()
	announcement := Announcement{
		ID:        m.nextID,
		Title:     title,
		Body:      body,
		AuthorID:  authorID,
		Priority:  priority,
		CreatedAt: now,
		UpdatedAt: now,
	}
	m.nextID++
	m.announcements = append(m.announcements, announcement)
	return announcement, nil
}

func (m *mockRepo) MarkAsRead(ctx context.Context, announcementID, memberID int32) error {
	// Check if already read
	for _, read := range m.reads {
		if read.AnnouncementID == announcementID && read.MemberID == memberID {
			return nil // Already read, no error
		}
	}
	
	// Add read record
	read := AnnouncementRead{
		AnnouncementID: announcementID,
		MemberID:       memberID,
		ReadAt:         time.Now(),
	}
	m.reads = append(m.reads, read)
	return nil
}

func (m *mockRepo) GetUnreadCount(ctx context.Context, memberID int32) (int, error) {
	count := 0
	for _, announcement := range m.announcements {
		isRead := false
		for _, read := range m.reads {
			if read.AnnouncementID == announcement.ID && read.MemberID == memberID {
				isRead = true
				break
			}
		}
		if !isRead {
			count++
		}
	}
	return count, nil
}

// ---- Helper functions ----

func setupRouter(repo Repo) *chi.Mux {
	r := chi.NewRouter()
	handlers := Handlers{Repo: repo}
	Mount(r, handlers)
	return r
}

// ---- Tests ----

func TestHandlers_List(t *testing.T) {
	authorID1 := int32(1)
	authorID2 := int32(2)

	repo := &mockRepo{
		announcements: []Announcement{
			{ID: 1, Title: "Urgent Notice", Body: "Important announcement", AuthorID: &authorID1, Priority: "urgent"},
			{ID: 2, Title: "Normal Update", Body: "Regular update", AuthorID: &authorID2, Priority: "normal"},
			{ID: 3, Title: "Low Priority", Body: "Less important info", AuthorID: nil, Priority: "low"},
		},
	}

	r := setupRouter(repo)

	// Test basic list without member context
	req := httptest.NewRequest("GET", "/announcements", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var announcements []AnnouncementWithReadStatus
	if err := json.Unmarshal(rr.Body.Bytes(), &announcements); err != nil {
		t.Fatal("failed to parse json:", err)
	}
	if len(announcements) != 3 {
		t.Errorf("expected 3 announcements, got %d", len(announcements))
	}

	// Test priority filter
	req = httptest.NewRequest("GET", "/announcements?priority=urgent", nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var urgentAnnouncements []AnnouncementWithReadStatus
	if err := json.Unmarshal(rr.Body.Bytes(), &urgentAnnouncements); err != nil {
		t.Fatal("failed to parse json:", err)
	}
	if len(urgentAnnouncements) != 1 {
		t.Errorf("expected 1 urgent announcement, got %d", len(urgentAnnouncements))
	}
	if urgentAnnouncements[0].Priority != "urgent" {
		t.Errorf("expected priority=urgent, got %s", urgentAnnouncements[0].Priority)
	}

	// Test with member context and read status
	req = httptest.NewRequest("GET", "/announcements?member_id=5", nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var announcementsWithReadStatus []AnnouncementWithReadStatus
	if err := json.Unmarshal(rr.Body.Bytes(), &announcementsWithReadStatus); err != nil {
		t.Fatal("failed to parse json:", err)
	}
	if len(announcementsWithReadStatus) != 3 {
		t.Errorf("expected 3 announcements, got %d", len(announcementsWithReadStatus))
	}
	// All should be unread initially
	for _, ann := range announcementsWithReadStatus {
		if ann.IsRead {
			t.Errorf("expected announcement %d to be unread", ann.ID)
		}
		if ann.MemberID != 5 {
			t.Errorf("expected member_id=5, got %d", ann.MemberID)
		}
	}
}

func TestHandlers_Create(t *testing.T) {
	repo := &mockRepo{}
	r := setupRouter(repo)

	tests := []struct {
		name           string
		payload        string
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:           "valid normal announcement",
			payload:        `{"title":"Test Announcement","body":"This is a test announcement","priority":"normal"}`,
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body []byte) {
				var announcement Announcement
				if err := json.Unmarshal(body, &announcement); err != nil {
					t.Fatal("failed to parse json:", err)
				}
				if announcement.Title != "Test Announcement" {
					t.Errorf("expected title='Test Announcement', got %s", announcement.Title)
				}
				if announcement.Priority != "normal" {
					t.Errorf("expected priority=normal, got %s", announcement.Priority)
				}
			},
		},
		{
			name:           "valid urgent announcement with author",
			payload:        `{"title":"Urgent Notice","body":"Very important","author_id":1,"priority":"urgent"}`,
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body []byte) {
				var announcement Announcement
				if err := json.Unmarshal(body, &announcement); err != nil {
					t.Fatal("failed to parse json:", err)
				}
				if announcement.Priority != "urgent" {
					t.Errorf("expected priority=urgent, got %s", announcement.Priority)
				}
				if announcement.AuthorID == nil || *announcement.AuthorID != 1 {
					t.Errorf("expected author_id=1, got %v", announcement.AuthorID)
				}
			},
		},
		{
			name:           "default priority when not specified",
			payload:        `{"title":"Default Priority","body":"Should default to normal"}`,
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body []byte) {
				var announcement Announcement
				if err := json.Unmarshal(body, &announcement); err != nil {
					t.Fatal("failed to parse json:", err)
				}
				if announcement.Priority != "normal" {
					t.Errorf("expected default priority=normal, got %s", announcement.Priority)
				}
			},
		},
		{
			name:           "missing title",
			payload:        `{"body":"No title provided","priority":"normal"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing body",
			payload:        `{"title":"No body","priority":"normal"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid priority",
			payload:        `{"title":"Test","body":"Test body","priority":"invalid"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid json",
			payload:        `{"title":"Test","body":}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/announcements", strings.NewReader(tt.payload))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.checkResponse != nil && rr.Code == http.StatusCreated {
				tt.checkResponse(t, rr.Body.Bytes())
			}
		})
	}
}

func TestHandlers_Get(t *testing.T) {
	authorID := int32(1)
	repo := &mockRepo{
		announcements: []Announcement{
			{ID: 1, Title: "Test Announcement", Body: "Test body", AuthorID: &authorID, Priority: "normal"},
		},
	}
	r := setupRouter(repo)

	// Test existing announcement
	req := httptest.NewRequest("GET", "/announcements/1", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var announcement AnnouncementWithReadStatus
	if err := json.Unmarshal(rr.Body.Bytes(), &announcement); err != nil {
		t.Fatal("failed to parse json:", err)
	}
	if announcement.ID != 1 {
		t.Errorf("expected id=1, got %d", announcement.ID)
	}

	// Test non-existent announcement
	req = httptest.NewRequest("GET", "/announcements/999", nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}

	// Test invalid id
	req = httptest.NewRequest("GET", "/announcements/invalid", nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestHandlers_MarkAsRead(t *testing.T) {
	authorID := int32(1)
	repo := &mockRepo{
		announcements: []Announcement{
			{ID: 1, Title: "Test Announcement", Body: "Test body", AuthorID: &authorID, Priority: "normal"},
		},
	}
	r := setupRouter(repo)

	// Test marking as read
	payload := `{"member_id":5}`
	req := httptest.NewRequest("POST", "/announcements/1/read", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var announcement AnnouncementWithReadStatus
	if err := json.Unmarshal(rr.Body.Bytes(), &announcement); err != nil {
		t.Fatal("failed to parse json:", err)
	}
	if !announcement.IsRead {
		t.Errorf("expected announcement to be marked as read")
	}
	if announcement.MemberID != 5 {
		t.Errorf("expected member_id=5, got %d", announcement.MemberID)
	}

	// Test marking non-existent announcement
	req = httptest.NewRequest("POST", "/announcements/999/read", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}

	// Test missing member_id
	payload = `{}`
	req = httptest.NewRequest("POST", "/announcements/1/read", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestHandlers_GetUnreadCount(t *testing.T) {
	authorID := int32(1)
	repo := &mockRepo{
		announcements: []Announcement{
			{ID: 1, Title: "Unread 1", Body: "Test", AuthorID: &authorID, Priority: "normal"},
			{ID: 2, Title: "Unread 2", Body: "Test", AuthorID: &authorID, Priority: "normal"},
			{ID: 3, Title: "Read", Body: "Test", AuthorID: &authorID, Priority: "normal"},
		},
		reads: []AnnouncementRead{
			{AnnouncementID: 3, MemberID: 5, ReadAt: time.Now()}, // Member 5 has read announcement 3
		},
	}
	r := setupRouter(repo)

	// Test unread count for member 5 (should have 2 unread)
	req := httptest.NewRequest("GET", "/announcements/unread?member_id=5", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var response struct {
		MemberID    int32 `json:"member_id"`
		UnreadCount int   `json:"unread_count"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal("failed to parse json:", err)
	}
	if response.MemberID != 5 {
		t.Errorf("expected member_id=5, got %d", response.MemberID)
	}
	if response.UnreadCount != 2 {
		t.Errorf("expected unread_count=2, got %d", response.UnreadCount)
	}

	// Test unread count for member 10 (should have 3 unread - none read)
	req = httptest.NewRequest("GET", "/announcements/unread?member_id=10", nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal("failed to parse json:", err)
	}
	if response.UnreadCount != 3 {
		t.Errorf("expected unread_count=3 for new member, got %d", response.UnreadCount)
	}

	// Test missing member_id parameter
	req = httptest.NewRequest("GET", "/announcements/unread", nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}

	// Test invalid member_id parameter
	req = httptest.NewRequest("GET", "/announcements/unread?member_id=invalid", nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
