# API Spec

Document REST endpoints and GraphQL (if used). Include error model and pagination.

## Proposals API

Base: /api/proposals

GET /              → 200 OK, list of Proposal
POST /             → 201 Created, Proposal
GET /{id}          → 200 OK, Proposal | 404 Not Found

Proposal:
{
  "id": 1,
  "title": "string",
  "body": "string",
  "status": "open",
  "created_at": "RFC3339 timestamp"
}
