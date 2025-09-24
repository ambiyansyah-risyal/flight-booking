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
- Start services: `docker compose up -d`
- Run migrations: `docker compose run --rm migrate up`
- Launch CLI: `go run ./cmd/flight-booking`

See `TASKS.md` for high-level scope and roadmap.
