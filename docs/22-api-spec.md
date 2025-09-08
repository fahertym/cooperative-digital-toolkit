# API Spec

## Conventions
- Base path: `/api`
- Auth header: `X-User-Id: <int>` required for protected routes
- Idempotency header (optional where supported): `X-Idempotency-Key: <string>`
- Content-Type: `application/json` for JSON, `text/csv` for CSV exports
- Timestamps: RFC3339 strings unless noted

## Headers

### Authentication
```
X-User-Id: 1
```
- Required on write routes for votes (POST/PUT), ledger create (POST), announcements read-state (POST)
- Returns `401` if missing or invalid

### Idempotency
```
X-Idempotency-Key: abc123
```
- Optional on POST `/api/ledger`
- If duplicate for the same user, server returns the original created resource (same JSON), success status consistent with implementation

---

## Health

### GET /healthz → 200 text/plain
Body: `ok`

---

## Proposals

### GET /api/proposals → 200
Query params:
- `limit` (int, optional, max 200)
- `offset` (int, optional)
Response headers (when provided): `X-Limit`, `X-Offset`
```json
[
  {"id":1,"title":"Bylaws update","body":"","status":"open","created_at":"2025-01-08T12:00:00Z"}
]
```

### POST /api/proposals → 201
Body: `{ "title": "...", "body": "..." }`
```json
{"id":1,"title":"Bylaws update","body":"","status":"open","created_at":"2025-01-08T12:00:00Z"}
```

### GET /api/proposals/{id} → 200 | 404

### POST /api/proposals/{id}/close → 200 | 404 | 409
Returns closed proposal object.

### GET /api/proposals/.csv → 200 text/csv
Note: CSV export returns all rows and ignores pagination parameters.

---

## Votes

Base: `/api/proposals/{id}/votes`

### POST /api/proposals/{id}/votes (auth) → 201 | 400 | 404 | 409
Body: `{ "choice": "for" | "against" | "abstain", "notes": "..." }`
```json
{"id":42,"proposal_id":1,"member_id":1,"choice":"for","notes":"","created_at":"2025-01-08T12:01:00Z"}
```

### PUT /api/proposals/{id}/votes (auth) → 200 | 400 | 404 | 409
Body: `{ "choice": "for" | "against" | "abstain", "notes": "..." }`

### GET /api/proposals/{id}/votes → 200
Query params:
- `limit` (int, optional, max 200)
- `offset` (int, optional)
Response headers (when provided): `X-Limit`, `X-Offset`
```json
[{"id":1,"proposal_id":1,"member_id":1,"choice":"for","notes":"","created_at":"2025-01-08T12:01:00Z"}]
```

### GET /api/proposals/{id}/votes/tally → 200 | 404
```json
{
  "proposal_id": 1,
  "status": "open",
  "total_eligible": 10,
  "votes_cast": 3,
  "quorum_met": false,
  "results": {"for":2, "against":1, "abstain":0},
  "outcome": "pending"
}
```

---

## Announcements

### GET /api/announcements → 200
Query params:
- `limit` (int, optional, max 200)
- `offset` (int, optional)
Response headers (when provided): `X-Limit`, `X-Offset`
```json
[
  {"id":1,"title":"Welcome","priority":"normal","created_at":"2025-01-08T12:00:00Z","is_read":false}
]
```

### POST /api/announcements → 201
Body: `{ "title":"...", "body":"...", "priority":"low|normal|high|urgent", "author_id": 1? }`
```json
{"id":1,"title":"Welcome","body":"...","priority":"normal","created_at":"2025-01-08T12:00:00Z","updated_at":"2025-01-08T12:00:00Z"}
```

### GET /api/announcements/{id} → 200 | 400 | 404
May include `is_read` and `member_id` when `member_id` query is provided.

### POST /api/announcements/{id}/read (auth) → 200 | 400 | 401 | 404
Returns the announcement with `is_read=true` for the current user.

### GET /api/announcements/unread?member_id={int} → 200 | 400
```json
{"member_id":1,"unread_count":0}
```

---

## Ledger

### POST /api/ledger (auth, idempotency optional) → 201 (or 200 on replay)
Body:
```json
{"type":"dues|contribution|expense|income","amount":50.00,"description":"...","notes":"..."}
```
Headers: `X-User-Id` required, `X-Idempotency-Key` optional
```json
{"id":7,"type":"dues","amount":50.00,"description":"...","member_id":1,"notes":"","created_at":"2025-01-08T12:03:00Z"}
```
Errors: `400` invalid input, `401` missing/invalid auth

### GET /api/ledger → 200
Query params:
- `type` (string, optional)
- `member_id` (int, optional)
- `limit` (int, optional, max 200)
- `offset` (int, optional)
Response headers (when provided): `X-Limit`, `X-Offset`

### GET /api/ledger/{id} → 200 | 400 | 404

### GET /api/ledger/.csv → 200 text/csv
Columns and order: `Date,Description,Type,Amount,Member ID,Notes,Reference`
Date is `YYYY-MM-DD` derived from `created_at`. Reference is the entry `id`.

---

## Errors and Content Types

Common statuses: `400` invalid input, `401` unauthorized, `404` not found, `409` conflict, `500` server error

Responses:
- JSON: `application/json`
- CSV exports: `text/csv; charset=utf-8`
- Health: `text/plain`

Error envelope (JSON):
```json
{"error":"message"}
```
All error responses from JSON endpoints use this shape. Examples:
- `400` invalid path param: `{ "error": "invalid id" }`
- `401` missing auth: `{ "error": "unauthorized" }`
- `404` missing resource: `{ "error": "not found" }`

Idempotency behavior:
- POST `/api/ledger` returns `201 Created` on first create and `200 OK` on idempotent replay for the same user and `X-Idempotency-Key`. Response body is the same resource JSON.
