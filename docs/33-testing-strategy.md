# Testing Strategy

## Backend
- **Unit tests**: handlers with mocked Repo interfaces.
- **Integration tests**: spin ephemeral Postgres (Docker) and test repo.
- **Table-driven cases**: clear inputs/outputs; golden files for JSON where helpful.

## Frontend
- **Vitest**: component logic, stores.
- **Playwright**: basic e2e for proposal create/list.

## Coverage Targets
- Short term: critical paths (proposals) ≥ 70%.
- Medium term: service surfaces ≥ 80%.

## Example: Handler test plan
- GET /api/proposals returns empty list → 200 []
- POST /api/proposals with no title → 400
- POST valid → 201; GET list shows new item

## Example: Handler Tests (Go)

We test HTTP endpoints with a mocked Repo and `httptest`, keeping DB out of unit tests.

```go
// See backend/internal/proposals/http_test.go for full code.
req := httptest.NewRequest("POST", "/api/proposals", strings.NewReader(`{"title":"Demo"}`))
req.Header.Set("Content-Type", "application/json")
rr := httptest.NewRecorder()
router.ServeHTTP(rr, req)
if rr.Code != http.StatusCreated { t.Fatalf("want 201 got %d", rr.Code) }
```

**Why this shape**

- Fast: no container spin-up for unit tests
- Deterministic: no flaky external services
- Focused: handler logic + routing + JSON contracts

**When to use real Postgres**

- For the repository layer (integration tests), spin a containerized Postgres and run against it to validate SQL. Keep those separate from unit tests so you can run unit tests frequently.

# Testing Strategy

- Unit, integration, e2e
- UAT with pilot co-ops
- Load & security tests
