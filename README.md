
# Cooperative Digital Toolkit

A member-owned, open-source platform for cooperative governance, finance, and collaboration.

## Quick Start (Dev)
- Backend: Go
- Frontend: Svelte
- DB: PostgreSQL
- Automations: n8n; Imports: Airbyte

```bash
make dev   # run backend + frontend in watch mode
make test  # run tests
```

Temporary auth: set `X-User-Id` header to a valid member id for write endpoints. Create a member via `POST /api/members` (role: `admin` or `member`).

## Docs
See `/docs` for strategy, product, and technical specifications.

## 📚 Documentation

All project docs live in [`/docs`](./docs). Start here:

- **Docs Home / Table of Contents:** [/docs/README.md](./docs/README.md)
- **Product Requirements (PRD):** [/docs/10-prd.md](./docs/10-prd.md)
- **Architecture Overview:** [/docs/20-architecture-overview.md](./docs/20-architecture-overview.md)
- **API Spec:** [/docs/22-api-spec.md](./docs/22-api-spec.md)
- **Contributor Guide:** [/docs/32-contributor-guide.md](./docs/32-contributor-guide.md)

## License
Apache-2.0
