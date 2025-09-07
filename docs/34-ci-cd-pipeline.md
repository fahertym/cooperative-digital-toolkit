# CI/CD Pipeline

## Current (v0.1)
- GitHub Actions on push/PR:
  - Backend: `go test ./...`
  - Frontend: `pnpm build`

### CI Jobs (GitHub Actions)

- **backend-tests**: runs `go test ./...` on PRs and pushes to main.
- **smoke** (manual): on the “Run workflow” button, set `run_smoke = true` to start Postgres, boot the server, and run `/scripts/smoke.sh` against `/healthz`, proposals CRUD, and CSV export.

## Near-Term Enhancements
- Linters: golangci-lint, eslint
- Security: CodeQL
- Artifacts: versioned releases
- Deploy: container image build + push (tags); prod job gated by manual approval

## Release Process
- Semantic versioning. Tag with `vX.Y.Z`.
- Changelog generated from Conventional Commits.

# CI/CD Pipeline

- Lint, test, build
- Versioning & changelog
- Release & rollback
