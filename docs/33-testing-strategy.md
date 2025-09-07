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

# Testing Strategy

- Unit, integration, e2e
- UAT with pilot co-ops
- Load & security tests
