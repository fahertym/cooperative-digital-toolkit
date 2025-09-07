# API Specification (v0.1)

Base URL (dev): `http://localhost:8080`

## Health
- `GET /healthz` → `200 OK` body: `ok`

## Proposals API
Base: `/api/proposals`

> See also:
> - **Domain guide:** [/docs/domains/proposals/README.md](./domains/proposals/README.md)
> - **Handler tests (reference):** [/backend/internal/proposals/http_test.go](../backend/internal/proposals/http_test.go)

> Domain guide: see [/docs/domains/proposals/README.md](./domains/proposals/README.md)

### List
- `GET /api/proposals`
- 200 OK
```json
[
  {
    "id": 2,
    "title": "Refactor check",
    "body": "now using repo layer",
    "status": "open",
    "created_at": "2025-09-07T20:40:56Z"
  }
]
```

### Create

* `POST /api/proposals`
* Body:

```json
{
  "title": "string (required)",
  "body": "string (optional)"
}
```

* 201 Created, returns Proposal.

### Get by ID

* `GET /api/proposals/{id}`
* 200 OK with Proposal | 404 Not Found

### Proposal schema
### Close
- `POST /api/proposals/{id}/close`
- Transitions `status: open → closed`.
- 200 OK with updated Proposal
- 404 Not Found if proposal does not exist
- 409 Conflict if proposal is not `open`

```json
{
  "id": 1,
  "title": "string",
  "body": "string",
  "status": "open|closed",
  "created_at": "RFC3339 timestamp"
}
```

### Export (CSV)
- `GET /api/proposals/.csv`
- Returns `text/csv` with columns: `id,title,body,status,created_at`
- Content-Disposition: `attachment; filename=proposals.csv`

## Auth (planned)

* WebAuthn bootstrap endpoints (Phase 2)
* Session cookies; CSRF protection (details TBD)

## Errors

* JSON error string in body when 4xx/5xx.
* Use `application/json` for all responses; UTF-8.
