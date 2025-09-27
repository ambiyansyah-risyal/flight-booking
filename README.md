# Flight Booking

[![Go CI](https://github.com/ambiyansyah-risyal/flight-booking/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/ambiyansyah-risyal/flight-booking/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/ambiyansyah-risyal/flight-booking/branch/main/graph/badge.svg)](https://codecov.io/gh/ambiyansyah-risyal/flight-booking)

Production-ready, scalable flight booking system built in Go. Designed with Clean Architecture and a CLI-first workflow to manage airports, routes, schedules, bookings, and seats. Uses PostgreSQL with sqlx, configuration via Viper, command wiring through Cobra, and automated migrations with Goose. CI runs on GitHub Actions with golangci-lint and Codecov coverage.

## Key Features
- Clean Architecture: decoupled domain, use cases, and adapters.
- Robust data layer: PostgreSQL + sqlx with context timeouts.
- CLI-first UX: Cobra commands for admin and passenger flows.
- Safe migrations: Goose as a one-shot migrator container.
- CI/CD ready: GitHub Actions, golangci-lint, Codecov coverage.

## Tech Stack
- Go 1.23+, Cobra, Viper, govalidator
- PostgreSQL, sqlx, Goose
- Docker, Docker Compose
- GitHub Actions, golangci-lint, Codecov

## Quick Start

### Prerequisites
- Go 1.23+, Docker, Docker Compose
- Make

### Available Commands
Run `make help` to see all available commands:
```
Flight Booking Makefile - Available targets:

  all -  Build the application (default target)
  build -  Build the application binary
  run -  Run the application
  test -  Run tests with race detection and coverage
  lint -  Run linter
  fmt -  Format code with gofmt and goimports
  mod-download -  Download go modules
  migrate-create -  Create a new database migration (usage: make migrate-create name=<migration_name>)
  migrate-up -  Run database migrations up
  migrate-down -  Run database migrations down
  migrate-redo -  Redo the last database migration
  migrate-status -  Show database migration status
  compose-up -  Start services with Docker Compose
  compose-down -  Stop and remove services with Docker Compose
  compose-migrate -  Run database migrations with Docker Compose
  compose-verify-all -  Run full verification using Docker Compose
  compose-verify -  Run basic verification workflow with Docker Compose
  test-e2e -  Run end-to-end tests
  local-verify -  Run local verification without Docker Compose
```

### Common Workflows
- One-shot verification (Compose): `make compose-verify` (builds images, starts DB, runs migrations, exercises basic commands, tears down)
- One-shot verification (Local): `make local-verify` (starts temporary Postgres, runs migrations, exercises CLI commands via `go run`)
- Full E2E test: `make compose-verify-all` (builds images, starts DB, runs migrations, exercises full workflows, tears down)

### Manual Setup
- Start DB: `make compose-up` (or `docker compose -f docker/compose.yml up -d db`)
- Run migrations: `make migrate-up` (or `docker compose -f docker/compose.yml run --rm migrate up`)
- Run CLI locally: `go run ./cmd/flight-booking` (reads `FLIGHT_*` env via Viper)

### Environment Variables
Example for local CLI:
```
FLIGHT_DB_HOST=localhost FLIGHT_DB_PORT=5432 FLIGHT_DB_USER=flight_app FLIGHT_DB_PASSWORD=app FLIGHT_DB_NAME=flight FLIGHT_DB_SSLMODE=disable
```

### Common CLI Commands
- Airports: `go run ./cmd/flight-booking airport list` | `create --code CGK --city Jakarta` | `update --code CGK --city NewName` | `delete CGK`
- DB health: `go run ./cmd/flight-booking db:ping`
- Bookings: `go run ./cmd/flight-booking booking search --origin CGK --destination SIN --date 2025-01-02` | `go run ./cmd/flight-booking booking book --schedule 1 --name "Alice"`

## End-to-End Test
- Requirements: Local Docker daemon available.
- Run: `make test-e2e`
  - Uses Testcontainers to start Postgres, applies Goose migrations, and exercises CLI commands (create/list/update/delete) against the DB.

See `TASKS.md` for high-level scope and roadmap.
