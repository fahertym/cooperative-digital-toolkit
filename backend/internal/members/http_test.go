package members

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "strconv"
    "strings"
    "testing"

    "github.com/go-chi/chi/v5"
)

// mockRepo implements Repo for handler tests
type mockRepo struct{
    byID map[int64]Member
    byEmail map[string]Member
    nextID int64
}

func (m *mockRepo) Create(_ context.Context, email, displayName, role string) (Member, error) {
    if m.byEmail == nil { m.byEmail = map[string]Member{} }
    if m.byID == nil { m.byID = map[int64]Member{} }
    if _, ok := m.byEmail[email]; ok { return Member{}, ErrConflict }
    if role == "" { role = "member" }
    m.nextID++
    mem := Member{ID: m.nextID, Email: email, DisplayName: displayName, Role: role}
    m.byEmail[email] = mem
    m.byID[mem.ID] = mem
    return mem, nil
}

func (m *mockRepo) GetByID(_ context.Context, id int64) (Member, error) {
    if v, ok := m.byID[id]; ok { return v, nil }
    return Member{}, ErrNotFound
}

func (m *mockRepo) GetByEmail(_ context.Context, email string) (Member, error) {
    if v, ok := m.byEmail[email]; ok { return v, nil }
    return Member{}, ErrNotFound
}

func setupRouter(repo Repo) *chi.Mux {
    r := chi.NewRouter()
    h := Handlers{Repo: repo}
    Mount(r, h)
    return r
}

func TestMembers_Create_ValidAndDuplicate(t *testing.T) {
    repo := &mockRepo{}
    r := setupRouter(repo)

    // valid create default role
    req := httptest.NewRequest("POST", "/members", strings.NewReader(`{"email":"a@ex.com","display_name":"A"}`))
    req.Header.Set("Content-Type", "application/json")
    rr := httptest.NewRecorder()
    r.ServeHTTP(rr, req)
    if rr.Code != http.StatusCreated { t.Fatalf("expected 201, got %d", rr.Code) }
    var m Member
    _ = json.Unmarshal(rr.Body.Bytes(), &m)
    if m.Role != "member" { t.Fatalf("expected role=member, got %s", m.Role) }

    // duplicate
    req = httptest.NewRequest("POST", "/members", strings.NewReader(`{"email":"a@ex.com","display_name":"A"}`))
    req.Header.Set("Content-Type", "application/json")
    rr = httptest.NewRecorder()
    r.ServeHTTP(rr, req)
    if rr.Code != http.StatusConflict { t.Fatalf("expected 409, got %d", rr.Code) }

    // admin create
    req = httptest.NewRequest("POST", "/members", strings.NewReader(`{"email":"admin@ex.com","display_name":"Admin","role":"admin"}`))
    req.Header.Set("Content-Type", "application/json")
    rr = httptest.NewRecorder()
    r.ServeHTTP(rr, req)
    if rr.Code != http.StatusCreated { t.Fatalf("expected 201, got %d", rr.Code) }
    _ = json.Unmarshal(rr.Body.Bytes(), &m)
    if m.Role != "admin" { t.Fatalf("expected role=admin, got %s", m.Role) }
}

func TestMembers_GetByID_And_FindByEmail(t *testing.T) {
    repo := &mockRepo{}
    // seed
    m, _ := repo.Create(context.Background(), "x@ex.com", "X", "member")
    r := setupRouter(repo)

    // by id
    req := httptest.NewRequest("GET", "/members/"+strconv.FormatInt(m.ID,10), nil)
    rr := httptest.NewRecorder()
    r.ServeHTTP(rr, req)
    if rr.Code != http.StatusOK { t.Fatalf("expected 200, got %d", rr.Code) }

    // by email
    req = httptest.NewRequest("GET", "/members?email=x@ex.com", nil)
    rr = httptest.NewRecorder()
    r.ServeHTTP(rr, req)
    if rr.Code != http.StatusOK { t.Fatalf("expected 200, got %d", rr.Code) }
}

