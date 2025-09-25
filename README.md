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
- Go 1.22+, Cobra, Viper, govalidator
- PostgreSQL, sqlx, Goose
- Docker, Docker Compose
- GitHub Actions, golangci-lint, Codecov

## Quick Start
- Prereqs: Go 1.22+, Docker, Docker Compose
 - One-shot verify (Compose): `make compose-verify` (builds images, starts DB, runs Goose migrations, pings DB, lists seeded airports, tears down).
 - One-shot verify (Local go run): `make local-verify` (spins a temporary Postgres with Docker, migrates with Goose CLI, and exercises all CLI commands via `go run`).
 - Full E2E (Compose): `make compose-verify-all` (builds images, starts DB, runs migrations, and exercises Airport and Airplane flows inside the app container, then tears down).
- Manual:
  - Start DB: `docker compose -f docker/compose.yml up -d db`
  - Migrate: `docker compose -f docker/compose.yml run --rm migrate up`
  - Run CLI locally: `go run ./cmd/flight-booking` (reads `FLIGHT_*` env via Viper)

Env examples for local CLI
- `FLIGHT_DB_HOST=localhost FLIGHT_DB_PORT=5432 FLIGHT_DB_USER=postgres FLIGHT_DB_PASSWORD=postgres FLIGHT_DB_NAME=flight FLIGHT_DB_SSLMODE=disable`

Common commands
- Airports: `go run ./cmd/flight-booking airport list` | `create --code CGK --city Jakarta` | `update --code CGK --city NewName` | `delete CGK`
- DB health: `go run ./cmd/flight-booking db:ping`
- Bookings: `go run ./cmd/flight-booking booking search --origin CGK --destination SIN --date 2025-01-02` | `go run ./cmd/flight-booking booking book --schedule 1 --name "Alice"`

## End-to-End Test
- Requirements: Local Docker daemon available.
- Run: `make test-e2e`
  - Uses Testcontainers to start Postgres, applies Goose migrations, and exercises CLI commands (create/list/update/delete) against the DB.

See `TASKS.md` for high-level scope and roadmap.
