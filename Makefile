
.SILENT: ;         # no noise
.DEFAULT_GOAL := help

help: ## Show help
	awk 'BEGIN {FS = ":.*##"; printf "\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

dev: ## Run backend and frontend in dev mode
	(cd backend && go run ./cmd/server) & (cd frontend && pnpm dev)

test: ## Run tests (backend + frontend)
	(cd backend && go test ./...) && (cd frontend && pnpm test)

lint: ## Lint (placeholder)
	echo "TODO: add golangci-lint and eslint"

build: ## Build backend and frontend
	(cd backend && go build -o ../bin/server ./cmd/server) && (cd frontend && pnpm build)

migrate: ## Apply embedded migrations by starting the server briefly
	( cd backend && DATABASE_URL=$${DATABASE_URL:-postgres://coop:coop@localhost:5432/coopdb?sslmode=disable} \
		go run ./cmd/server & pid=$$!; sleep 2; kill $$pid || true )
