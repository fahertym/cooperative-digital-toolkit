package proposals

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
