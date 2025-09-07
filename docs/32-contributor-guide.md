# Contributor Guide

## Local Dev (Ubuntu/WSL)
```bash
# Go
GO_VERSION=1.22.7
wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc && source ~/.bashrc

# Node
curl -fsSL https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash
source ~/.nvm/nvm.sh && nvm install --lts && corepack enable

# deps
cd backend && go mod tidy && cd ..
cd frontend && pnpm install && cd ..

# DB
docker compose up -d
make dev  # runs backend+frontend
```

## Branching & Commits

* Trunk-based with short-lived branches.
* Conventional Commits: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`.

## PR Checklist

* Tests for new behavior
* Docs updated (API/spec/user stories)
* CI green

## Code Style

* Keep handlers thin; domain/repo in `internal/<domain>`.
* Avoid global state; use contexts.
* Use interfaces to enable testing/mocking.

# Contributor Guide

- Dev setup
- Coding standards
- Branching: trunk-based w/ short-lived feature branches (or GitFlow)
- PR checklist
