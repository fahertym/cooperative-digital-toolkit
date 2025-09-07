# Security & Identity

## Today
- CORS locked to frontend origin in dev
- JSON-only API; strict content types
- RBAC roles: admin, member (expand later)
- TLS via reverse proxy in production

## Roadmap
- WebAuthn for passwordless auth (Phase 2)
- Sessions with HttpOnly, Secure cookies; CSRF tokens where needed
- DID/VC integration for portable membership credentials

## Threat Model (initial)
- Common web risks (XSS, CSRF, SQLi) → mitigated by headers, parameterized queries, escaping, CSRF protection
- Privilege escalation → RBAC + event audits
- Data exfiltration → minimal PII stored; encryption at rest (DB-level/infra)
