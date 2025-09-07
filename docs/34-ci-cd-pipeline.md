# CI/CD Pipeline

## Current (v0.1)
- GitHub Actions on push/PR:
  - Backend: `go test ./...`
  - Frontend: `pnpm build`

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
