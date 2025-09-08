# Architecture Overview

## Approach
- Start with a **modular monolith** to reduce operational cost/complexity.
- Extract services only when real usage metrics show pressure (throughput, team scale, fault domains).
- Everything critical is event-logged for audits and democratic accountability.

## Components
- **Backend:** Go 1.22, chi router, pgx pool, CORS.
- **Frontend:** Svelte + Vite (SvelteKit upgrade likely).
- **Database:** PostgreSQL 16 (CockroachDB path for federation).
- **Automations:** n8n; **Ingest:** Airbyte.
- **Identity:** Phase 1 WebAuthn; Phase 2 DID/VC (DIDKit, Aries).
- **Containerization:** docker-compose for dev; Kubernetes optional later.

## Data model (initial)
- `proposals(id, title, body, status, created_at)`
- Future: `members`, `organizations`, `votes`, `ledger_entries`, `events`.

## Event logging
- Each mutating action (create proposal, post announcement, ledger entry) emits an event with actor, timestamp, payload hash.

## Offline-first
- Browser cache (IndexedDB); reconcile queued mutations with server.
- Conflict policy: app-level rule, user prompt, or “last write wins” depending on entity.

## Offline and Idempotency

### Client queue
- The client can queue creates while offline
- Each queued request includes an idempotency key stored locally

### Server behavior
- For endpoints that support idempotency (ledger today), requests include `X-Idempotency-Key`
- The server ensures one logical create per `(member_id, key)`
- Policy: On conflict, return the original resource with success

### Conflicts
- Duplicate votes are prevented by a DB uniqueness constraint on `(proposal_id, member_id)`
- Announcements read-state uses uniqueness on `(announcement_id, member_id)`

## Security
- TLS terminated by reverse proxy in prod.
- Role-based access control (RBAC) — roles: admin, member (more later).
- Minimal attack surface: small deps, clear CORS, JSON-only, strict headers.

## Scale & Federation Plan
- Single DB per tenant at first.
- Introduce read replicas if needed.
- Federation: identity portability; claim exchange between co-ops; selective data commons.

## Deployment Sketch
- **Dev:** docker-compose for Postgres + Adminer; `make dev` for servers.
- **Prod (small):** single VM or container host with Postgres + app process.
- **Prod (bigger):** managed Postgres or Cockroach; replicas; object storage for attachments.

# Architecture Overview

- Modular monolith → microservices (strangler pattern)
- Event log for auditability
- Offline-first sync model
