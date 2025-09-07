package proposals

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

// ---- Mock Repo ----

type mockRepo struct {
	items  []Proposal
	nextID int32
}

func (m *mockRepo) List(_ context.Context) ([]Proposal, error) {
	// Return a copy to avoid mutation by callers
	out := make([]Proposal, len(m.items))
	copy(out, m.items)
	return out, nil
}

func (m *mockRepo) Get(_ context.Context, id int32) (Proposal, error) {
	for _, p := range m.items {
		if p.ID == id {
			return p, nil
		}
	}
	return Proposal{}, ErrNotFound
}

func (m *mockRepo) Create(_ context.Context, title, body string) (Proposal, error) {
	if m.nextID == 0 {
		m.nextID = 1
	}
	p := Proposal{
		ID:     m.nextID,
		Title:  title,
		Body:   body,
		Status: "open",
		// CreatedAt left zero; handler tests don't assert it
	}
	m.nextID++
	// prepend newest
	m.items = append([]Proposal{p}, m.items...)
	return p, nil
}

func (m *mockRepo) Close(_ context.Context, id int32) (Proposal, error) {
	for i, p := range m.items {
		if p.ID == id {
			if p.Status != "open" {
				return Proposal{}, ErrConflict
			}
			p.Status = "closed"
			m.items[i] = p
			return p, nil
		}
	}
	return Proposal{}, ErrNotFound
}

// ---- Test Router Setup ----

func testRouter(repo Repo) http.Handler {
	r := chi.NewRouter()
	h := Handlers{Repo: repo}
	r.Route("/api", func(api chi.Router) {
		Mount(api, h)
	})
	return r
}

// ---- Tests ----

func TestListEmpty(t *testing.T) {
	repo := &mockRepo{}
	r := testRouter(repo)

	req := httptest.NewRequest("GET", "/api/proposals", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var got []Proposal
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty list, got %d items", len(got))
	}
}

func TestCreateRequiresTitle(t *testing.T) {
	repo := &mockRepo{}
	r := testRouter(repo)

	req := httptest.NewRequest("POST", "/api/proposals", strings.NewReader(`{"body":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing title, got %d", rr.Code)
	}
}

func TestCreateAndList(t *testing.T) {
	repo := &mockRepo{}
	r := testRouter(repo)

	body := `{"title":"Demo","body":"Hello"}`
	req := httptest.NewRequest("POST", "/api/proposals", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (body=%s)", rr.Code, rr.Body.String())
	}

	var created Proposal
	if err := json.Unmarshal(rr.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal created: %v", err)
	}
	if created.ID == 0 || created.Title != "Demo" || created.Status != "open" {
		t.Fatalf("unexpected created: %+v", created)
	}

	// List should show 1 item
	req2 := httptest.NewRequest("GET", "/api/proposals", nil)
	rr2 := httptest.NewRecorder()
	r.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr2.Code)
	}
	var list []Proposal
	if err := json.Unmarshal(rr2.Body.Bytes(), &list); err != nil {
		t.Fatalf("unmarshal list: %v", err)
	}
	if len(list) != 1 || list[0].ID != created.ID {
		t.Fatalf("unexpected list: %+v", list)
	}
}

func TestGetNotFound(t *testing.T) {
	repo := &mockRepo{}
	r := testRouter(repo)

	req := httptest.NewRequest("GET", "/api/proposals/999", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestCloseHappyPath(t *testing.T) {
	repo := &mockRepo{}
	r := testRouter(repo)

	// Create open proposal
	reqC := httptest.NewRequest("POST", "/api/proposals", strings.NewReader(`{"title":"Close me"}`))
	reqC.Header.Set("Content-Type", "application/json")
	rrC := httptest.NewRecorder()
	r.ServeHTTP(rrC, reqC)
	if rrC.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rrC.Code)
	}
	var created Proposal
	_ = json.Unmarshal(rrC.Body.Bytes(), &created)

	// Close it
	req := httptest.NewRequest("POST", "/api/proposals/"+itoa(created.ID)+"/close", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", rr.Code, rr.Body.String())
	}
	var closed Proposal
	_ = json.Unmarshal(rr.Body.Bytes(), &closed)
	if closed.Status != "closed" {
		t.Fatalf("expected status=closed, got %s", closed.Status)
	}
}

func TestCloseAlreadyClosed(t *testing.T) {
	repo := &mockRepo{}
	r := testRouter(repo)

	// Create
	reqC := httptest.NewRequest("POST", "/api/proposals", strings.NewReader(`{"title":"Twice"}`))
	reqC.Header.Set("Content-Type", "application/json")
	rrC := httptest.NewRecorder()
	r.ServeHTTP(rrC, reqC)
	var created Proposal
	_ = json.Unmarshal(rrC.Body.Bytes(), &created)

	// Close once
	req1 := httptest.NewRequest("POST", "/api/proposals/"+itoa(created.ID)+"/close", nil)
	rr1 := httptest.NewRecorder()
	r.ServeHTTP(rr1, req1)
	if rr1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr1.Code)
	}

	// Close again -> 409
	req2 := httptest.NewRequest("POST", "/api/proposals/"+itoa(created.ID)+"/close", nil)
	rr2 := httptest.NewRecorder()
	r.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rr2.Code)
	}
}

func TestExportCSV(t *testing.T) {
	repo := &mockRepo{}
	r := testRouter(repo)

	// Seed two proposals
	for _, ttl := range []string{"CSV One", "CSV Two"} {
		req := httptest.NewRequest("POST", "/api/proposals", strings.NewReader(`{"title":"`+ttl+`"}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		if rr.Code != http.StatusCreated {
			t.Fatalf("seed create: want 201 got %d", rr.Code)
		}
	}

	// Export
	req := httptest.NewRequest("GET", "/api/proposals/.csv", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("csv: want 200 got %d", rr.Code)
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/csv") {
		t.Fatalf("csv: unexpected content-type %q", ct)
	}
	body := rr.Body.String()
	if !strings.HasPrefix(body, "id,title,body,status,created_at") {
		t.Fatalf("csv header mismatch: %q", body[:min(60, len(body))])
	}
	if !strings.Contains(body, "CSV One") || !strings.Contains(body, "CSV Two") {
		t.Fatalf("csv missing rows:\n%s", body)
	}
}

func itoa(v int32) string { return strconv.FormatInt(int64(v), 10) }
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
