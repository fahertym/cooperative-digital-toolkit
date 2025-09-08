# Security and Identity

## Temporary session model for MVP
- Header `X-User-Id` is required on protected routes (votes POST/PUT, ledger POST, announcements read-state POST)
- Middleware validates the header and injects `member_id` into request context
- Invalid or missing header returns `401 Unauthorized`

## Planned upgrade path
- Replace header-based session with WebAuthn or passwordless email link
- Server-side session store or tokens with rotation
- Migrate protected routes without changing domain semantics

## Authorization notes
- Write endpoints affected: POST votes, PUT votes, POST ledger, POST announcements/{id}/read
- Read endpoints may enrich responses with per-member `is_read` flags when a user is present

## Idempotency for resilience
- `X-Idempotency-Key` supported on POST `/api/ledger`
- Uniqueness enforced per `(member_id, idempotency_key)` when key is present
- On repeat, server returns the original entry with success
