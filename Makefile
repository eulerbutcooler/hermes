.PHONY: help infra-up infra-down db-migrate-up db-migrate-down db-migrate-create db-reset db-shell db-status setup dev-core dev-hooks dev-worker build

# Database connection
DB_USER := user
DB_PASSWORD := password
DB_NAME := hermes
DB_HOST := localhost
DB_PORT := 5432
DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
POSTGRES_CONTAINER := hermes-postgres
MIGRATIONS_PATH := services/hermes-core/db/migrations

# Colors
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## Infrastructure commands

infra-up: ## Start Postgres + NATS in Docker
	@echo "$(YELLOW)Starting infrastructure...$(NC)"
	@docker compose up -d
	@echo "$(YELLOW)Waiting for services to be healthy...$(NC)"
	@sleep 3
	@docker compose ps
	@echo "$(GREEN)✓ Infrastructure ready!$(NC)"

infra-down: ## Stop infrastructure
	@docker compose down

infra-logs: ## View infrastructure logs
	@docker compose logs -f

infra-clean: ## Stop and remove all data (WARNING: deletes volumes!)
	@echo "$(RED)⚠️  This will delete all database data!$(NC)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		docker compose down -v; \
		echo "$(GREEN)✓ Cleaned!$(NC)"; \
	fi

## Database migration commands (using migrate CLI)

db-migrate-up: ## Run all pending migrations
	@echo "$(YELLOW)Running migrations...$(NC)"
	@migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" up
	@echo "$(GREEN)✓ Migrations applied!$(NC)"

db-migrate-down: ## Rollback last migration
	@echo "$(YELLOW)Rolling back last migration...$(NC)"
	@migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down 1
	@echo "$(GREEN)✓ Rollback complete!$(NC)"

db-migrate-force: ## Force migration version (use: make db-migrate-force VERSION=1)
	@if [ -z "$(VERSION)" ]; then \
		echo "$(RED)Error: VERSION not specified. Usage: make db-migrate-force VERSION=1$(NC)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Forcing migration to version $(VERSION)...$(NC)"
	@migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" force $(VERSION)
	@echo "$(GREEN)✓ Forced to version $(VERSION)$(NC)"

db-migrate-version: ## Show current migration version
	@migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" version

db-migrate-create: ## Create a new migration (use: make db-migrate-create NAME=add_users_table)
	@if [ -z "$(NAME)" ]; then \
		echo "$(RED)Error: NAME not specified. Usage: make db-migrate-create NAME=add_users_table$(NC)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Creating new migration: $(NAME)$(NC)"
	@migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $(NAME)
	@echo "$(GREEN)✓ Created migration files in $(MIGRATIONS_PATH)$(NC)"

db-reset: infra-up ## Reset database (drop all tables and re-run migrations)
	@echo "$(YELLOW)Resetting database...$(NC)"
	@migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" drop -f
	@$(MAKE) db-migrate-up
	@echo "$(GREEN)✓ Database reset complete!$(NC)"

db-shell: ## Open psql shell in Postgres container
	@docker exec -it $(POSTGRES_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME)

db-status: ## Show database tables and row counts
	@echo "$(YELLOW)Database tables:$(NC)"
	@docker exec -i $(POSTGRES_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "\dt"
	@echo ""
	@echo "$(YELLOW)Row counts:$(NC)"
	@docker exec -i $(POSTGRES_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) -c "\
		SELECT 'users' as table, count(*) FROM users \
		UNION ALL SELECT 'relays', count(*) FROM relays \
		UNION ALL SELECT 'relay_actions', count(*) FROM relay_actions \
		UNION ALL SELECT 'execution_logs', count(*) FROM execution_logs;"

## Development commands

dev-core: ## Run hermes-core API server
	@echo "$(YELLOW)Starting hermes-core...$(NC)"
	@cd services/hermes-core && go run cmd/api/main.go

dev-hooks: ## Run hermes-hooks webhook server
	@echo "$(YELLOW)Starting hermes-hooks...$(NC)"
	@cd services/hermes-hooks && go run cmd/server/main.go

dev-worker: ## Run hermes-worker background processor
	@echo "$(YELLOW)Starting hermes-worker...$(NC)"
	@cd services/hermes-worker && go run cmd/main.go

## Build commands

build: ## Build all services into bin/ directory
	@echo "$(YELLOW)Building all services...$(NC)"
	@mkdir -p bin
	@go build -o bin/hermes-core ./services/hermes-core/cmd/api
	@go build -o bin/hermes-hooks ./services/hermes-hooks/cmd/server
	@go build -o bin/hermes-worker ./services/hermes-worker/cmd
	@echo "$(GREEN)✓ Built binaries in bin/$(NC)"
	@ls -lh bin/

clean: ## Clean build artifacts
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	@rm -rf bin/
	@find . -name "*.log" -type f -delete
	@echo "$(GREEN)✓ Cleaned!$(NC)"

## Setup commands

setup: infra-up db-migrate-up ## Complete first-time setup (infra + migrations)
	@echo ""
	@echo "$(GREEN)✓✓✓ Setup complete! ✓✓✓$(NC)"
	@echo ""
	@echo "$(YELLOW)Next steps:$(NC)"
	@echo "  1. Terminal 1: $(GREEN)make dev-core$(NC)   (API server on :3000)"
	@echo "  2. Terminal 2: $(GREEN)make dev-hooks$(NC)  (Webhook ingestion on :8080)"
	@echo "  3. Terminal 3: $(GREEN)make dev-worker$(NC) (Background workers)"
	@echo ""
	@echo "$(YELLOW)Test the flow:$(NC)"
	@echo "  $(GREEN)curl -X POST http://localhost:8080/hooks/test-relay -d '{\"test\":\"data\"}'$(NC)"
	@echo ""
	@echo "$(YELLOW)Database commands:$(NC)"
	@echo "  $(GREEN)make db-status$(NC)  - View tables and counts"
	@echo "  $(GREEN)make db-shell$(NC)   - Open psql shell"

check: ## Check if all infrastructure is healthy
	@echo "$(YELLOW)Checking infrastructure health...$(NC)"
	@docker exec $(POSTGRES_CONTAINER) pg_isready -U $(DB_USER) -d $(DB_NAME) > /dev/null 2>&1 && \
		echo "$(GREEN)✓ Postgres is healthy$(NC)" || echo "$(RED)✗ Postgres not ready$(NC)"
	@curl -sf http://localhost:8222/healthz > /dev/null 2>&1 && \
		echo "$(GREEN)✓ NATS is healthy$(NC)" || echo "$(RED)✗ NATS not ready$(NC)"
	@migrate -version > /dev/null 2>&1 && \
		echo "$(GREEN)✓ migrate CLI installed$(NC)" || echo "$(RED)✗ migrate CLI not found (run: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest)$(NC)"

## Additional utilities

test: ## Run all tests
	@echo "$(YELLOW)Running tests...$(NC)"
	@go test -v ./...

lint: ## Run linter (requires golangci-lint)
	@echo "$(YELLOW)Running linter...$(NC)"
	@golangci-lint run ./...

.DEFAULT_GOAL := help
