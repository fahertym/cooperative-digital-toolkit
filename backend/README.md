# Backend

Go HTTP server exposing `/healthz` and `/api/*` domains.

Auth: temporary header `X-User-Id` conveys a numeric member id. The server loads the member and attaches it to request context. If missing, requests are treated as `guest` for read-only endpoints. Write endpoints require authentication and may require specific roles.

New in this PR:
- Members domain: `POST /api/members`, `GET /api/members/{id}`, `GET /api/members?email=`
- AuthZ rules: `POST /api/announcements` requires `admin`. `POST /api/announcements/{id}/read` requires authentication.

Quick dev run:
- `DATABASE_URL=postgres://coop:coop@localhost:5432/coopdb?sslmode=disable go run ./cmd/server`
- `curl -s localhost:8080/healthz`
