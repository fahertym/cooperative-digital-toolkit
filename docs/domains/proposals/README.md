# Proposals Domain — Functional Guide (v0.1)

This guide explains how proposals work for product, design, and engineering.

## Purpose

Enable members to read and act on governance items; enable admins to run compliant, auditable decision flows with minimal overhead.

## Entities

- **Proposal**: title, body, status (`open|closed|archived`), created_at.
- (Phase 2) **Vote**: member choice on a proposal (yes/no/abstain) with quorum rules.
- (Planned) **Event**: audit entries for creation/closure/exports.

## Lifecycle

1. **Create** (admin): title/body → status=`open`.
2. **Read** (member): visible in lists.
3. **Close** (admin) — `POST /api/proposals/{id}/close` (Implemented): status=`closed`. After closure:

   - Proposal becomes read-only.
   - Future vote endpoints should reject writes.
4. **Archive** (admin, later): status=`archived` to hide from default list.

### Status Rules

- Allowed: `open`, `closed`, `archived`.
- Transitions:
  - `open → closed` (allowed)
  - `open → archived` (allowed)
  - `closed → archived` (allowed)
  - `closed → open` (not allowed)
- Enforce in API and (eventually) via DB `CHECK`.

## Access Control (MVP)

- **Admin**: create, close (later), export.
- **Member**: read.
- Later: per-org roles; ACLs per proposal if needed.

## API (current)

- `GET /api/proposals` → list newest first.
- `POST /api/proposals {title, body?}` → create (status=`open`).
- `GET /api/proposals/{id}` → fetch by id.
- `POST /api/proposals/{id}/close` → transition to `closed` (only from `open`).

See details in [`/docs/22-api-spec.md`](../../22-api-spec.md).

## Validation

- `title` required; recommend 1–200 chars.
- `body` optional (sanitize or render as plaintext initially).

## Pagination

- Current: simple `ORDER BY id DESC`.
- Future: keyset pagination: `WHERE id < last_seen_id ORDER BY id DESC LIMIT N`.

## Errors

- `400` invalid json or missing title.
- `404` proposal not found.
- `500` DB errors.

## Audit/Events (planned)

Emit events for:

- `proposal.create`
- `proposal.close`
- `proposal.export` (CSV/PDF)
  - Include: actor, entity id, timestamp, payload hash or diff.

## Exports
- **CSV**: `GET /api/proposals/.csv` → columns: id, title, body, status, created_at
- Used for quick reporting and downstream accounting/records.

## UI Notes

- Creation form: title (required), body (markdown/plain).
- List: show `#id`, title, status chip, created_at (local time).
- Detail: render body safely; show actions based on role and status.
- Accessibility: keyboard focus order; ensure readable contrast.

#### Admin toggle & Close action (dev UI)

- **Admin mode toggle:** In the header, a checkbox toggles an “admin” flag stored in `localStorage` (`isAdmin = "1"`). This is a **dev convenience only** and will be replaced by real auth (WebAuthn + session) later.
  - Enable via UI or in console: `localStorage.setItem('isAdmin','1')`
  - Disable: `localStorage.removeItem('isAdmin')`

- **Status chip:** Each proposal shows a chip reflecting `status` (`open`, `closed`, `archived`).

- **Close button (admin-only):** When Admin mode is on and `status === 'open'`, a Close button appears per proposal.
  - Action: `POST /api/proposals/{id}/close`
  - Success: returns updated Proposal with `status: "closed"`; button disappears.
  - Errors: `404` if not found, `409` if not `open` (already closed/archived).
  - Accessibility: the button label is “Close”; styled as a destructive action. Consider adding a confirm dialog in production.

- **Source:** `frontend/src/main.js` (`closeProposal(id)`, admin toggle, rendering).

## Performance

- Add index on `(created_at DESC)` and `(status)` (already recommended).
- Consider partial indexes: `WHERE status='open'` for default list views.

## Telemetry (nice-to-have later)

- Count proposal views per org (privacy-aware).
- Time-to-first-view from creation.
- Export usage counts.

## Examples

**Create**

```bash
curl -s -X POST http://localhost:8080/api/proposals \
  -H 'Content-Type: application/json' \
  -d '{"title":"Adopt shared purchasing policy","body":"Bulk buying across co-ops"}'
```

**List**

```bash
curl -s http://localhost:8080/api/proposals | jq .
```

**Get**

```bash
curl -s http://localhost:8080/api/proposals/1 | jq .
```

## Roadmap Links

- Data model: [`/docs/23-data-models.md`](../../23-data-models.md)
- Architecture: [`/docs/20-architecture-overview.md`](../../20-architecture-overview.md)
- User stories: [`/docs/12-user-stories.md`](../../12-user-stories.md)
