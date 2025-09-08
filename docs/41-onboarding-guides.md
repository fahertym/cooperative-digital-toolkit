# Onboarding and Quick Start

## Run backend
```bash
cd backend
go run ./cmd/server
# exports PORT and DATABASE_URL as needed
```

## Headers to know

- `X-User-Id: 1` for protected routes (votes create/update, ledger create, announcements read)
- `X-Idempotency-Key: abc123` for ledger POST idempotency

## Smoke test

```bash
PORT=8080 BASE="http://localhost:$PORT" ./scripts/smoke.sh
```

Smoke covers:

- Proposals create → vote → tally → close
- Ledger create with idempotency → CSV
- Announcements create → read → unread
