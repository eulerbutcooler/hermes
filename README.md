# hermes

# Infrastructure
make infra-up          # Start Postgres + NATS
make infra-down        # Stop infrastructure
make infra-logs        # View container logs

# Database
make db-migrate-up     # Run migrations
make db-migrate-down   # Rollback last migration
make db-reset          # Drop all & re-migrate
make db-status         # Show tables & counts
make db-shell          # Open psql

# Development
make dev-core          # Run API server
make dev-hooks         # Run webhook ingestion
make dev-worker        # Run background workers

# Build & Deploy
make build             # Build all binaries
make check             # Health check all services
make setup             # First-time setup

# Utilities
make help              # Show all commands
