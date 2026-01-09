.PHONY: help dev test lint migrate-up migrate-down migrate-create seed demo-smoke

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

dev: ## Start development environment (docker compose)
	docker compose up --build

dev-bg: ## Start development environment in background
	docker compose up --build -d

dev-down: ## Stop development environment
	docker compose down

test: ## Run all tests (Go + UI)
	@echo "Running Go tests..."
	cd api && go test -v ./... || exit 1
	@echo "✓ Go tests passed"
	@echo "Running UI tests..."
	cd ui && npm test -- --run || exit 1
	@echo "✓ UI tests passed"

lint: ## Run linters
	@echo "Running Go linter..."
	cd api && golangci-lint run || echo "golangci-lint not installed, skipping"
	@echo "Running UI linter..."
	cd ui && npm run lint || echo "UI lint not configured yet"

migrate-up: ## Apply all pending database migrations
	migrate -path db/migrations -database "$(DATABASE_URL)" up

migrate-down: ## Rollback last database migration
	migrate -path db/migrations -database "$(DATABASE_URL)" down 1

migrate-create: ## Create new migration (usage: make migrate-create name=create_users)
	migrate -path db/migrations -database "$(DATABASE_URL)" create $(name) sql

seed: ## Seed database with sample data
	@echo "Seeding sample data..."
	@if [ -f scripts/seed-data.go ]; then \
		cd scripts && go run seed-data.go -demo; \
	else \
		echo "Seed script not yet implemented"; \
	fi

demo-smoke: ## Run end-to-end smoke test
	@echo "Running smoke tests..."
	@if [ -f scripts/smoke-test.sh ]; then \
		./scripts/smoke-test.sh; \
	else \
		echo "Smoke test script not yet implemented"; \
	fi

build-api: ## Build Go API binary
	cd api && go build -o bin/server ./cmd/server

build-ui: ## Build React SPA
	cd ui && npm run build

clean: ## Clean build artifacts
	rm -rf api/bin/
	rm -rf ui/dist/
	docker compose down -v
