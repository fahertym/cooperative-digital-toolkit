package ledger

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"coop.tools/backend/internal/httpmw"
	"github.com/go-chi/chi/v5"
)

// ---- Mock Repo ----

type mockRepo struct {
    entries []LedgerEntry
    nextID  int32
}

func (m *mockRepo) List(_ context.Context, filters *ListFilters) ([]LedgerEntry, error) {
	// Return a copy to avoid mutation by callers
	out := make([]LedgerEntry, 0)

	for _, entry := range m.entries {
		include := true

		// Apply filters
		if filters != nil {
			if filters.Type != "" && entry.Type != filters.Type {
				include = false
			}
			if filters.MemberID != nil {
				if entry.MemberID == nil || *entry.MemberID != *filters.MemberID {
					include = false
				}
			}
			// Skip date filters in mock for simplicity
		}

		if include {
			out = append(out, entry)
		}
	}

	return out, nil
}

func (m *mockRepo) Get(_ context.Context, id int32) (LedgerEntry, error) {
	for _, entry := range m.entries {
		if entry.ID == id {
			return entry, nil
		}
	}
	return LedgerEntry{}, ErrNotFound
}

func (m *mockRepo) Create(_ context.Context, entryType, description string, amount float64, memberID *int32, notes string, idempotencyKey string) (LedgerEntry, bool, error) {
    if m.nextID == 0 {
        m.nextID = 1
    }
    entry := LedgerEntry{
        ID:          m.nextID,
        Type:        entryType,
        Amount:      amount,
        Description: description,
        MemberID:    memberID,
        Notes:       notes,
        CreatedAt:   time.Now(),
    }
    m.nextID++
    m.entries = append(m.entries, entry)
    return entry, false, nil
}

// ---- Helper functions ----

func setupRouter(repo Repo) *chi.Mux {
    r := chi.NewRouter()
    r.Use(httpmw.WithAuth(func(ctx context.Context, id int64) (httpmw.Principal, bool, error) {
        if id <= 0 { return httpmw.Principal{}, false, nil }
        return httpmw.Principal{MemberID: id, Role: "member"}, true, nil
    }))
    handlers := Handlers{Repo: repo}
    Mount(r, handlers)
    return r
}

// ---- Tests ----

func TestHandlers_List(t *testing.T) {
	memberID1 := int32(1)
	memberID2 := int32(2)

	repo := &mockRepo{
		entries: []LedgerEntry{
			{ID: 1, Type: "dues", Amount: 50.00, Description: "Monthly dues", MemberID: &memberID1},
			{ID: 2, Type: "expense", Amount: -25.50, Description: "Office supplies", MemberID: nil},
			{ID: 3, Type: "contribution", Amount: 100.00, Description: "Annual contribution", MemberID: &memberID2},
		},
	}

	r := setupRouter(repo)

	// Test basic list
	req := httptest.NewRequest("GET", "/ledger", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var entries []LedgerEntry
	if err := json.Unmarshal(rr.Body.Bytes(), &entries); err != nil {
		t.Fatal("failed to parse json:", err)
	}
	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}

	// Test type filter
	req = httptest.NewRequest("GET", "/ledger?type=dues", nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var filteredEntries []LedgerEntry
	if err := json.Unmarshal(rr.Body.Bytes(), &filteredEntries); err != nil {
		t.Fatal("failed to parse json:", err)
	}
	if len(filteredEntries) != 1 {
		t.Errorf("expected 1 entry with type=dues, got %d", len(filteredEntries))
	}
	if filteredEntries[0].Type != "dues" {
		t.Errorf("expected type=dues, got %s", filteredEntries[0].Type)
	}
}

func TestHandlers_Create(t *testing.T) {
	repo := &mockRepo{}
	r := setupRouter(repo)

	tests := []struct {
		name           string
		payload        string
		headers        map[string]string
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:           "valid dues entry",
			payload:        `{"type":"dues","amount":50.00,"description":"Monthly dues","member_id":1}`,
			headers:        map[string]string{"X-User-Id": "1", "Content-Type": "application/json"},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body []byte) {
				var entry LedgerEntry
				if err := json.Unmarshal(body, &entry); err != nil {
					t.Fatal("failed to parse json:", err)
				}
				if entry.Type != "dues" {
					t.Errorf("expected type=dues, got %s", entry.Type)
				}
				if entry.Amount != 50.00 {
					t.Errorf("expected amount=50.00, got %f", entry.Amount)
				}
				if entry.Description != "Monthly dues" {
					t.Errorf("expected description='Monthly dues', got %s", entry.Description)
				}
			},
		},
		{
			name:           "missing auth",
			payload:        `{"type":"dues","amount":50.00,"description":"Monthly dues"}`,
			headers:        map[string]string{"Content-Type": "application/json"},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid type",
			payload:        `{"type":"invalid","amount":50.00,"description":"Test"}`,
			headers:        map[string]string{"X-User-Id": "1", "Content-Type": "application/json"},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/ledger", strings.NewReader(tt.payload))
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
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
	memberID := int32(1)
	repo := &mockRepo{
		entries: []LedgerEntry{
			{ID: 1, Type: "dues", Amount: 50.00, Description: "Monthly dues", MemberID: &memberID},
		},
	}
	r := setupRouter(repo)

	// Test existing entry
	req := httptest.NewRequest("GET", "/ledger/1", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var entry LedgerEntry
	if err := json.Unmarshal(rr.Body.Bytes(), &entry); err != nil {
		t.Fatal("failed to parse json:", err)
	}
	if entry.ID != 1 {
		t.Errorf("expected id=1, got %d", entry.ID)
	}

	// Test non-existent entry
	req = httptest.NewRequest("GET", "/ledger/999", nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}

	// Test invalid id
	req = httptest.NewRequest("GET", "/ledger/invalid", nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestHandlers_ExportCSV(t *testing.T) {
	memberID1 := int32(1)
	memberID2 := int32(2)

	repo := &mockRepo{
		entries: []LedgerEntry{
			{ID: 1, Type: "dues", Amount: 50.00, Description: "Monthly dues", MemberID: &memberID1, CreatedAt: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)},
			{ID: 2, Type: "expense", Amount: -25.50, Description: "Office supplies", MemberID: nil, Notes: "For new office", CreatedAt: time.Date(2025, 1, 2, 14, 30, 0, 0, time.UTC)},
			{ID: 3, Type: "contribution", Amount: 100.00, Description: "Annual contribution", MemberID: &memberID2, CreatedAt: time.Date(2025, 1, 3, 9, 15, 0, 0, time.UTC)},
		},
	}

	r := setupRouter(repo)

	req := httptest.NewRequest("GET", "/ledger/.csv", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "text/csv; charset=utf-8" {
		t.Errorf("expected csv content type, got %s", contentType)
	}

	disposition := rr.Header().Get("Content-Disposition")
	if !strings.Contains(disposition, "attachment") || !strings.Contains(disposition, "ledger.csv") {
		t.Errorf("expected attachment disposition with ledger.csv, got %s", disposition)
	}

	csv := rr.Body.String()
	lines := strings.Split(strings.TrimSpace(csv), "\n")

	// Check header
	if len(lines) < 1 {
		t.Fatal("expected at least header line")
	}
	expectedHeader := "Date,Description,Type,Amount,Member ID,Notes,Reference"
	if lines[0] != expectedHeader {
		t.Errorf("expected header '%s', got '%s'", expectedHeader, lines[0])
	}

	// Check data rows
	if len(lines) != 4 { // header + 3 entries
		t.Errorf("expected 4 lines (header + 3 entries), got %d", len(lines))
	}

	// Check first data row
	expectedFirstRow := "2025-01-01,Monthly dues,dues,50.00,1,,1"
	if lines[1] != expectedFirstRow {
		t.Errorf("expected first row '%s', got '%s'", expectedFirstRow, lines[1])
	}

	// Check expense row (negative amount, no member ID, has notes)
	expectedExpenseRow := "2025-01-02,Office supplies,expense,-25.50,,For new office,2"
	if lines[2] != expectedExpenseRow {
		t.Errorf("expected expense row '%s', got '%s'", expectedExpenseRow, lines[2])
	}
}
