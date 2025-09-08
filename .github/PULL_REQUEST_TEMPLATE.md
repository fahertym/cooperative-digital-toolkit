### What changed
- [ ] Code
- [ ] Docs: 22-api-spec.md
- [ ] Docs: 23-data-models.md
- [ ] Docs: 24-security-identity.md
- [ ] Docs: 20-architecture-overview.md
- [ ] Docs: 41-onboarding-guides.md
- [ ] Smoke script updated

### Checks
- [ ] Unit tests `go test ./...` pass
- [ ] Smoke `scripts/smoke.sh` passes locally
- [ ] CSV headers and order verified by tests
- [ ] Protected routes return 401 without X-User-Id
- [ ] Idempotency for ledger validated by test

## Summary
Explain the change and the user impact.

## Linked Issues
- Closes #

## Type
- [ ] feat
- [ ] fix
- [ ] docs
- [ ] refactor
- [ ] test
- [ ] chore

## Checklists

### Code
- [ ] Tests added/updated (unit/integration as appropriate)
- [ ] CI green

### Docs (required for product or API changes)
- [ ] PRD updated if requirements changed: [/docs/10-prd.md](../docs/10-prd.md)
- [ ] API Spec updated if endpoints changed: [/docs/22-api-spec.md](../docs/22-api-spec.md)
- [ ] Architecture/docs updated if design changed: [/docs](../docs)
