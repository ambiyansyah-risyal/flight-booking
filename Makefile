SHELL := /bin/bash

# Help command to list all available targets
help: ## Display this help message
help: ## Display this help message
	@echo "Flight Booking Makefile - Available targets:"
	@echo
	@fgrep -h "##" Makefile | grep -v "help:" | sed "s/:.*##/ - /" | sed "s/^/  /"


# Config
APP := flight-booking
PKG := ./...
GO := go
GOFLAGS :=

# Database
DATABASE_URL ?= postgres://flight_migrator:migrator@localhost:5432/flight?sslmode=disable
MIGRATIONS_DIR := migrations
GOOSE := $$(shell command -v goose 2>/dev/null)

.PHONY: all build run test lint fmt mod-download help
all: build ## Build the application (default target)

build: ## Build the application binary
	@VERSION=$$(git describe --tags --abbrev=0 2>/dev/null || echo dev); \\
	COMMIT=$$(git rev-parse --short HEAD); \\
	DATE=$$(date -u +%Y-%m-%dT%H:%M:%SZ); \\
	LDFLAGS=\"-s -w -X github.com/ambiyansyah-risyal/flight-booking/internal/adapter/cli.Version=$$VERSION -X github.com/ambiyansyah-risyal/flight-booking/internal/adapter/cli.Commit=$$COMMIT -X github.com/ambiyansyah-risyal/flight-booking/internal/adapter/cli.BuildDate=$$DATE\"; \\
	$$(GO) build $$(GOFLAGS) -ldflags \"$$LDFLAGS\" -o bin/$$(APP) ./cmd/flight-booking

run: ## Run the application
	$$(GO) run ./cmd/flight-booking

test: ## Run tests with race detection and coverage
	$$(GO) test $$(PKG) -race -coverprofile=coverage.out -covermode=atomic

lint: ## Run linter
	golangci-lint run

fmt: ## Format code with gofmt and goimports
	gofmt -s -w .
	goimports -w .

mod-download: ## Download go modules
	$$(GO) mod download

# Migrations (Goose)
.PHONY: migrate-create migrate-up migrate-down migrate-redo migrate-status
migrate-create: ## Create a new database migration (usage: make migrate-create name=<migration_name>)
 ifndef name
	$(error Usage: make migrate-create name=<migration_name>)
 endif
	@if [ -z \"$$(GOOSE)\" ]; then \\
	GO111MODULE=on $$(GO) install github.com/pressly/goose/v3/cmd/goose@latest ; \\
	fi
	goose -dir $$(MIGRATIONS_DIR) create $$(name) sql

migrate-up: ## Run database migrations up
	@if [ -z \"$$(GOOSE)\" ]; then \\
	GO111MODULE=on $$(GO) install github.com/pressly/goose/v3/cmd/goose@latest ; \\
	fi
	goose -dir $$(MIGRATIONS_DIR) postgres \"$$(DATABASE_URL)\" up

migrate-down: ## Run database migrations down
	@if [ -z \"$$(GOOSE)\" ]; then \\
	GO111MODULE=on $$(GO) install github.com/pressly/goose/v3/cmd/goose@latest ; \\
	fi
	goose -dir $$(MIGRATIONS_DIR) postgres \"$$(DATABASE_URL)\" down

migrate-redo: ## Redo the last database migration
	@if [ -z \"$$(GOOSE)\" ]; then \\
	GO111MODULE=on $$(GO) install github.com/pressly/goose/v3/cmd/goose@latest ; \\
	fi
	goose -dir $$(MIGRATIONS_DIR) postgres \"$$(DATABASE_URL)\" redo

migrate-status: ## Show database migration status
	@if [ -z \"$$(GOOSE)\" ]; then \\
	GO111MODULE=on $$(GO) install github.com/pressly/goose/v3/cmd/goose@latest ; \\
	fi
	goose -dir $$(MIGRATIONS_DIR) postgres \"$$(DATABASE_URL)\" status

# Docker Compose helpers
.PHONY: compose-up compose-down compose-migrate compose-verify-all compose-verify
compose-up: ## Start services with Docker Compose
	docker compose -f docker/compose.yml up -d --build

compose-down: ## Stop and remove services with Docker Compose
	docker compose -f docker/compose.yml down -v

compose-migrate: ## Run database migrations with Docker Compose
	docker compose -f docker/compose.yml run --rm migrate up

compose-verify-all: ## Run full verification using Docker Compose
	bash scripts/verify/compose-verify.sh

compose-verify: ## Run basic verification workflow with Docker Compose
	# Build images
	docker compose -f docker/compose.yml build app migrate
	# Start DB only
	docker compose -f docker/compose.yml up -d db
	# Wait for DB readiness
	@echo \"Waiting for DB to be ready...\"
	@until docker compose -f docker/compose.yml exec -T db pg_isready -U postgres -d flight >/dev/null 2>&1; do \\
		echo -n \".\"; sleep 1; \\
	done; echo
	# Run migrations, then basic checks
	docker compose -f docker/compose.yml run --rm migrate up
	docker compose -f docker/compose.yml run --rm app db:ping
	docker compose -f docker/compose.yml run --rm app airport list --limit 10
	# Tear down
	docker compose -f docker/compose.yml down -v

.PHONY: test-e2e local-verify
test-e2e: ## Run end-to-end tests
	go test -tags e2e ./e2e -v

local-verify: ## Run local verification without Docker Compose
	bash scripts/verify/local-verify.sh