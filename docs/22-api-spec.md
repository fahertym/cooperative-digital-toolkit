# API Specification (v0.1)

Base URL (dev): `http://localhost:8080`

All JSON endpoints use UTF-8 encoding. Error responses return JSON string in body for 4xx/5xx codes.

## Health

### GET /healthz
Returns service health status.

**Response:**
- `200 OK` - Service healthy
- Body: `ok` (plain text)

## Proposals API

Base: `/api/proposals`

> See also:
> - **Domain guide:** [/docs/domains/proposals/README.md](./domains/proposals/README.md)  
> - **Handler tests (reference):** [/backend/internal/proposals/http_test.go](../backend/internal/proposals/http_test.go)

### Proposal Schema
```json
{
  "id": "integer (int32, auto-generated)",
  "title": "string (required, max length TBD)",
  "body": "string (optional)",
  "status": "string (enum: 'open', 'closed')",
  "created_at": "string (RFC3339 timestamp in UTC)"
}
```

### GET /api/proposals
List all proposals in reverse chronological order.

**Response:**
- `200 OK` - Array of proposal objects
- `500 Internal Server Error` - Database or server error

**Example Response:**
```json
[
  {
    "id": 2,
    "title": "Refactor check",
    "body": "now using repo layer", 
    "status": "open",
    "created_at": "2025-01-07T20:40:56Z"
  },
  {
    "id": 1,
    "title": "Approve budget",
    "body": "",
    "status": "closed", 
    "created_at": "2025-01-06T15:20:30Z"
  }
]
```

### POST /api/proposals
Create a new proposal.

**Request Body:**
```json
{
  "title": "string (required)",
  "body": "string (optional, defaults to empty string)"
}
```

**Response:**
- `201 Created` - Returns created proposal
- `400 Bad Request` - Invalid JSON or missing title  
- `500 Internal Server Error` - Database or server error

**Example Request:**
```json
{
  "title": "Approve new member handbook",
  "body": "Updated handbook includes remote work policies and conflict resolution procedures."
}
```

**Example Response:**
```json
{
  "id": 3,
  "title": "Approve new member handbook",
  "body": "Updated handbook includes remote work policies and conflict resolution procedures.",
  "status": "open",
  "created_at": "2025-01-08T10:15:00Z"
}
```

### GET /api/proposals/{id}
Get a specific proposal by ID.

**Path Parameters:**
- `id` (integer) - Proposal ID

**Response:**
- `200 OK` - Returns proposal object
- `404 Not Found` - Proposal does not exist
- `500 Internal Server Error` - Database or server error

### POST /api/proposals/{id}/close
Close an open proposal (transitions status from 'open' to 'closed').

**Path Parameters:**
- `id` (integer) - Proposal ID

**Response:**
- `200 OK` - Returns updated proposal with status='closed'
- `404 Not Found` - Proposal does not exist
- `409 Conflict` - Proposal is already closed or not in 'open' status
- `500 Internal Server Error` - Database or server error

**Example Response:**
```json
{
  "id": 1,
  "title": "Approve budget",
  "body": "FY2025 operating budget",
  "status": "closed",
  "created_at": "2025-01-06T15:20:30Z"
}
```

### GET /api/proposals/.csv
Export all proposals as CSV file.

**Response:**
- `200 OK` - CSV file download
- `500 Internal Server Error` - Database or server error

**Headers:**
- `Content-Type: text/csv; charset=utf-8`
- `Content-Disposition: attachment; filename=proposals.csv`

**CSV Format:**
```csv
id,title,body,status,created_at
1,"Approve budget","FY2025 operating budget",closed,2025-01-06T15:20:30Z
2,"Refactor check","now using repo layer",open,2025-01-07T20:40:56Z
```

## Voting API

Base: `/api/proposals/{id}/votes`

> **Status:** ✅ **Implemented and working**

### POST /api/proposals/{id}/votes
Cast a vote on an open proposal.

**Path Parameters:**
- `id` (integer) - Proposal ID

**Request Body:**
```json
{
  "member_id": "integer (required)",
  "choice": "string (enum: 'for', 'against', 'abstain')",
  "notes": "string (optional, member's reasoning)"
}
```

**Response:**
- `201 Created` - Vote successfully cast
- `400 Bad Request` - Invalid choice or missing member_id
- `404 Not Found` - Proposal does not exist  
- `409 Conflict` - Proposal is closed, or member already voted
- `500 Internal Server Error` - Database or server error

**Example Request:**
```json
{
  "member_id": 1,
  "choice": "for",
  "notes": "I support this proposal"
}
```

**Example Response:**
```json
{
  "id": 1,
  "proposal_id": 12,
  "member_id": 1,
  "choice": "for",
  "notes": "I support this proposal",
  "created_at": "2025-01-08T10:15:00Z"
}
```

### PUT /api/proposals/{id}/votes
Update an existing vote on an open proposal.

**Path Parameters:**
- `id` (integer) - Proposal ID

**Request Body:**
```json
{
  "member_id": "integer (required)",
  "choice": "string (enum: 'for', 'against', 'abstain')",
  "notes": "string (optional, member's reasoning)"
}
```

**Response:**
- `200 OK` - Vote successfully updated
- `400 Bad Request` - Invalid choice or missing member_id
- `404 Not Found` - Proposal or vote does not exist  
- `409 Conflict` - Proposal is closed
- `500 Internal Server Error` - Database or server error

### GET /api/proposals/{id}/votes
List all votes for a proposal.

**Path Parameters:**
- `id` (integer) - Proposal ID

**Response:**
- `200 OK` - Array of vote objects
- `404 Not Found` - Proposal does not exist
- `500 Internal Server Error` - Database or server error

**Example Response:**
```json
[
  {
    "id": 1,
    "proposal_id": 12,
    "member_id": 1,
    "choice": "for",
    "notes": "I support this proposal",
    "created_at": "2025-01-08T10:15:00Z"
  },
  {
    "id": 2,
    "proposal_id": 12,
    "member_id": 2,
    "choice": "against",
    "notes": "I have concerns about this",
    "created_at": "2025-01-08T10:16:00Z"
  }
]
```

### GET /api/proposals/{id}/votes/tally
Get vote tally and results for a proposal.

**Path Parameters:**
- `id` (integer) - Proposal ID

**Response:**
- `200 OK` - Vote tally object
- `404 Not Found` - Proposal does not exist
- `500 Internal Server Error` - Database or server error

**Example Response:**
```json
{
  "proposal_id": 12,
  "status": "open",
  "total_eligible": 10,
  "votes_cast": 3,
  "quorum_met": false,
  "results": {
    "for": 2,
    "against": 1,
    "abstain": 0
  },
  "outcome": "pending"
}
```

## Ledger API (Planned - Phase 2)

Base: `/api/ledger`

> **Status:** Not yet implemented. Part of MVP Phase 2 expansion.

### Ledger Entry Schema
```json
{
  "id": "integer (auto-generated)",
  "type": "string (enum: 'dues', 'contribution', 'expense', 'income')",
  "amount": "number (decimal, in dollars)",
  "description": "string (required)",
  "member_id": "integer (optional, null for org-level entries)",
  "notes": "string (optional)",
  "created_at": "string (RFC3339 timestamp)"
}
```

### GET /api/ledger
List ledger entries with optional filtering.

**Query Parameters:**
- `type` (optional) - Filter by entry type
- `member_id` (optional) - Filter by member
- `from_date` (optional) - Start date (RFC3339)
- `to_date` (optional) - End date (RFC3339)

### POST /api/ledger  
Create a new ledger entry.

### GET /api/ledger/.csv
Export ledger entries as CSV (QuickBooks/Xero compatible format).

## Announcements API (Planned - Phase 2)

Base: `/api/announcements`

> **Status:** Not yet implemented. Part of MVP Phase 2 expansion.

### Announcement Schema
```json
{
  "id": "integer (auto-generated)",
  "title": "string (required)",
  "body": "string (required, markdown supported)",
  "author_id": "integer (member who created)",
  "priority": "string (enum: 'low', 'normal', 'high', 'urgent')",
  "created_at": "string (RFC3339 timestamp)",
  "updated_at": "string (RFC3339 timestamp)"
}
```

### GET /api/announcements
List announcements with read status per authenticated member.

### POST /api/announcements
Create a new announcement (admin/authorized members only).

### POST /api/announcements/{id}/read
Mark announcement as read for authenticated member.

## Authentication API (Planned - Phase 2)

Base: `/api/auth`

> **Status:** Not yet implemented. Will replace dev "Admin mode" toggle.
> 
> **Strategy:** WebAuthn for passwordless authentication or email magic links.

### POST /api/auth/register
Register new member account.

### POST /api/auth/login  
Initiate login flow (email link or WebAuthn challenge).

### POST /api/auth/webauthn/begin
Begin WebAuthn authentication ceremony.

### POST /api/auth/webauthn/finish
Complete WebAuthn authentication ceremony.

### POST /api/auth/logout
End authenticated session.

## Error Handling

All API endpoints return errors in consistent JSON format:

**4xx Client Errors:**
```json
"error message describing the issue"
```

**5xx Server Errors:**
```json  
"internal server error occurred"
```

**Common HTTP Status Codes:**
- `400 Bad Request` - Invalid input, malformed JSON, missing required fields
- `401 Unauthorized` - Authentication required (future, with auth implementation)
- `403 Forbidden` - Insufficient permissions (future, with auth implementation)  
- `404 Not Found` - Resource does not exist
- `409 Conflict` - Operation conflicts with current state (e.g., already closed)
- `500 Internal Server Error` - Database connection, server issues

## Content Types

**Request Headers:**
- `Content-Type: application/json` (for POST requests with JSON body)
- `Accept: application/json` (recommended for JSON responses)
- `Accept: text/csv` (for CSV export endpoints)

**Response Headers:**  
- `Content-Type: application/json` (JSON responses)
- `Content-Type: text/csv; charset=utf-8` (CSV exports)
- `Content-Type: text/plain` (health check only)

## Rate Limiting (Future)

Rate limiting is not currently implemented but will be added in post-MVP phases:
- Per-IP limits for unauthenticated endpoints
- Per-member limits for authenticated operations  
- Burst allowances for interactive use cases
