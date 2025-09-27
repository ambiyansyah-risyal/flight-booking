SHELL := /bin/bash

# Config
APP := flight-booking
PKG := ./...
GO := go
GOFLAGS :=

# Database
DATABASE_URL ?= postgres://flight_migrator:migrator@localhost:5432/flight?sslmode=disable
MIGRATIONS_DIR := migrations
GOOSE := $(shell command -v goose 2>/dev/null)

.PHONY: all build run test lint fmt tidy vendor mod-download
all: build

build:
	@VERSION=$$(git describe --tags --abbrev=0 2>/dev/null || echo dev); \
	COMMIT=$$(git rev-parse --short HEAD); \
	DATE=$$(date -u +%Y-%m-%dT%H:%M:%SZ); \
	LDFLAGS="-s -w -X github.com/ambiyansyah-risyal/flight-booking/internal/adapter/cli.Version=$$VERSION -X github.com/ambiyansyah-risyal/flight-booking/internal/adapter/cli.Commit=$$COMMIT -X github.com/ambiyansyah-risyal/flight-booking/internal/adapter/cli.BuildDate=$$DATE"; \
	$(GO) build $(GOFLAGS) -ldflags "$$LDFLAGS" -o bin/$(APP) ./cmd/flight-booking

run:
	$(GO) run ./cmd/flight-booking

test:
	$(GO) test $(PKG) -race -coverprofile=coverage.out -covermode=atomic

lint:
	golangci-lint run

fmt:
	gofmt -s -w .
	goimports -w .

mod-download:
	$(GO) mod download

# Migrations (Goose)
.PHONY: migrate-create migrate-up migrate-down migrate-redo migrate-status
migrate-create:
	ifndef name
		$(error Usage: make migrate-create name=<migration_name>)
	endif
	@if [ -z "$(GOOSE)" ]; then \
		GO111MODULE=on $(GO) install github.com/pressly/goose/v3/cmd/goose@latest ; \
	fi
	goose -dir $(MIGRATIONS_DIR) create $(name) sql

migrate-up:
	@if [ -z "$(GOOSE)" ]; then \
		GO111MODULE=on $(GO) install github.com/pressly/goose/v3/cmd/goose@latest ; \
	fi
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" up

migrate-down:
	@if [ -z "$(GOOSE)" ]; then \
		GO111MODULE=on $(GO) install github.com/pressly/goose/v3/cmd/goose@latest ; \
	fi
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" down

migrate-redo:
	@if [ -z "$(GOOSE)" ]; then \
		GO111MODULE=on $(GO) install github.com/pressly/goose/v3/cmd/goose@latest ; \
	fi
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" redo

migrate-status:
	@if [ -z "$(GOOSE)" ]; then \
		GO111MODULE=on $(GO) install github.com/pressly/goose/v3/cmd/goose@latest ; \
	fi
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" status

# Docker Compose helpers (run from repo root)
.PHONY: compose-up compose-down compose-migrate
compose-up:
	docker compose -f docker/compose.yml up -d --build

compose-down:
	docker compose -f docker/compose.yml down -v

compose-migrate:
	docker compose -f docker/compose.yml run --rm migrate up

.PHONY: compose-verify-all
compose-verify-all:
	bash scripts/verify/compose-verify.sh

.PHONY: compose-verify
compose-verify:
	# Build images
	docker compose -f docker/compose.yml build app migrate
	# Start DB only
	docker compose -f docker/compose.yml up -d db
	# Wait for DB readiness
	@echo "Waiting for DB to be ready..."
	@until docker compose -f docker/compose.yml exec -T db pg_isready -U postgres -d flight >/dev/null 2>&1; do \
		echo -n "."; sleep 1; \
	done; echo
	# Run migrations, then basic checks
	docker compose -f docker/compose.yml run --rm migrate up
	docker compose -f docker/compose.yml run --rm app db:ping
	docker compose -f docker/compose.yml run --rm app airport list --limit 10
	# Tear down
	docker compose -f docker/compose.yml down -v

.PHONY: test-e2e
test-e2e:
	go test -tags e2e ./e2e -v

.PHONY: local-verify
local-verify:
	bash scripts/verify/local-verify.sh
