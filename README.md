# Hermes

Hermes is an automation platform designed to connect different services and execute automated workflows via webhooks and cron schedules.

## 🏗 Architecture

The Hermes platform is built with a microservices architecture:

* **`hermes-core`**: The primary backend REST API built in Go. It handles the management of users, workflows (relays), encrypted secrets, OAuth connections, and execution history.
* **`hermes-hooks`**: A fast webhook ingestion service that receives external HTTP requests and queues them onto a NATS JetStream event bus.
* **`hermes-worker`**: A background job processor that consumes events from the queue, resolves secrets, manages OAuth token refreshes, and executes the actions defined in a workflow. It also features a cron scheduler for time-based triggers.
* **`web`**: A modern Next.js frontend dashboard used to construct workflows, manage connections, and monitor execution logs.
* **`hermes-common`**: A shared Go library containing common utilities such as structured logging and encryption.

## 🔌 Supported Integrations

The execution engine currently supports the following actions:
* Discord Webhooks
* Slack Webhooks
* Email (via authenticated Google and Microsoft OAuth connections)
* HTTP Requests
* Debug Logging

## 🚀 Prerequisites

Ensure you have the following installed on your system before starting:

* Docker and Docker Compose (for PostgreSQL and NATS)
* Go 1.25.6 or higher
* Node.js (for the Next.js frontend)
* `make` utility
* `golang-migrate` CLI (for database migrations)

## 🛠 Getting Started

### 1. Environment Configuration
Copy the `.env.example` file to create your local `.env` file. Fill in the required variables, including `ENCRYPTION_KEY`, `JWT_SECRET`, and any OAuth Client IDs you plan to use.

### 2. Infrastructure Setup
Start the local infrastructure (PostgreSQL and NATS) and run the initial database migrations:
```bash
make setup

# Terminal 1: Start the core API server (runs on port 3000)
make dev-core

# Terminal 2: Start the webhook ingestion server (runs on port 8080)
make dev-hooks

# Terminal 3: Start the background execution worker
make dev-worker
```

cd web
npm install
npm run dev

The root Makefile provides several utilities for building and managing the application:
```bash
    make infra-up: Starts Postgres and NATS via Docker Compose.

    make infra-down: Stops the local infrastructure.

    make infra-logs: Views infrastructure container logs.

    make db-migrate-up: Runs all pending database migrations.

    make db-reset: Drops all database tables and re-runs migrations.

    make db-status: Displays the current database tables and row counts.

    make build: Builds all Go service binaries and outputs them into the bin/ directory.

    make check: Verifies that all infrastructure components are healthy and ready.
```
