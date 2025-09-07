# Tech Stack Rationale

## Backend: Go
- Fast, memory-efficient, small binaries, great concurrency.
- Easy to ship self-contained services with minimal ops overhead.

## Router: chi
- Lightweight, idiomatic, composable middlewares.

## Database: PostgreSQL
- Mature, reliable, ubiquitous; easy local dev; CockroachDB wire-compat path.

## Frontend: Svelte (+ Vite)
- Small bundles; simpler reactivity; low bandwidth friendly.
- Option: upgrade to SvelteKit for routing + server rendering.

## Automations: n8n; Ingest: Airbyte
- Bridge to existing accounting/CRM/POS without building one-off glue.
- Graphical, accessible to non-developers; exportable JSON workflows.

## Identity: WebAuthn → DID/VC
- Start with passkeys to eliminate password risk.
- Evolve to portable, verifiable membership credentials for federations.

## CI/CD
- GitHub Actions: test, build. Later: linters, CodeQL, release tagging.

## Licensing
- Apache-2.0: permissive, encourages ecosystem growth, reduces lock-in.

# Tech Stack Rationale

- Backend: Go
- Frontend: Svelte
- DB: PostgreSQL → CockroachDB (Phase 2)
- Automations: n8n; data ingest: Airbyte
- Identity: WebAuthn → DIDKit/VCs
