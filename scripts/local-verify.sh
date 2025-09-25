#!/usr/bin/env bash
set -euo pipefail

# Local verification script without Docker Compose.
# - Starts a temporary Postgres container
# - Bootstraps roles (flight_migrator, flight_app)
# - Runs Goose migrations
# - Exercises CLI via `go run` (version, db:ping, airport CRUD)

NAME=${NAME:-pg_cli_check}
PORT=${PORT:-6546}
DB=${DB:-flight}

cleanup() {
  docker rm -f "$NAME" >/dev/null 2>&1 || true
}
trap cleanup EXIT

echo "[1/6] Starting Postgres on :$PORT"
cleanup
docker run --rm -d --name "$NAME" -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB="$DB" -p "$PORT:5432" postgres:16-alpine >/dev/null

echo -n "[2/6] Waiting for Postgres to be ready"
for i in {1..60}; do
  if docker exec "$NAME" pg_isready -U postgres -d "$DB" >/dev/null 2>&1; then break; fi
  printf "."; sleep 1
  if [ "$i" -eq 60 ]; then echo "\nERROR: Postgres did not become ready" >&2; exit 1; fi
done
echo

echo "[3/6] Bootstrapping roles and grants (migrator/app)"
docker exec -i "$NAME" psql -U postgres -d "$DB" <<'SQL'
DO $$
BEGIN
   IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'flight_migrator') THEN
      CREATE ROLE flight_migrator LOGIN PASSWORD 'migrator';
   END IF;
   IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'flight_app') THEN
      CREATE ROLE flight_app LOGIN PASSWORD 'app';
   END IF;
END$$;
GRANT ALL PRIVILEGES ON DATABASE flight TO flight_migrator;
GRANT USAGE, CREATE ON SCHEMA public TO flight_migrator;
GRANT CONNECT ON DATABASE flight TO flight_app;
GRANT USAGE ON SCHEMA public TO flight_app;
ALTER DEFAULT PRIVILEGES FOR ROLE flight_migrator IN SCHEMA public
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO flight_app;
ALTER DEFAULT PRIVILEGES FOR ROLE flight_migrator IN SCHEMA public
    GRANT USAGE, SELECT ON SEQUENCES TO flight_app;
SQL

echo "[4/6] Running Goose migrations as migrator"
GOOSE_BIN="$(go env GOPATH)/bin/goose"
if [ ! -x "$GOOSE_BIN" ]; then
  echo "Installing goose CLI..."
  go install github.com/pressly/goose/v3/cmd/goose@latest
fi
DBURL="postgres://flight_migrator:migrator@localhost:$PORT/$DB?sslmode=disable"
"$GOOSE_BIN" -dir migrations postgres "$DBURL" up

echo "[5/6] Exporting app env and exercising CLI with go run"
export FLIGHT_DB_HOST=localhost
export FLIGHT_DB_PORT=$PORT
export FLIGHT_DB_USER=flight_app
export FLIGHT_DB_PASSWORD=app
export FLIGHT_DB_NAME=$DB
export FLIGHT_DB_SSLMODE=disable

set -x
go run ./cmd/flight-booking version
go run ./cmd/flight-booking db:ping
go run ./cmd/flight-booking airport list --limit 10
go run ./cmd/flight-booking airport create --code TST --city "Test City"
go run ./cmd/flight-booking airport list --limit 10
go run ./cmd/flight-booking airport update --code TST --city "New Test City"
go run ./cmd/flight-booking airport list --limit 10
go run ./cmd/flight-booking airport delete TST
go run ./cmd/flight-booking airport list --limit 10
set +x

echo "[6/6] Done. Container will be removed."
