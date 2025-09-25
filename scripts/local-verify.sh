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

# Helpers for asserting outcomes
run_ok() {
  echo ">>> OK: $*"
  set -x; "$@"; rc=$?; set +x
  if [ $rc -ne 0 ]; then echo "Expected OK but failed: $*" >&2; exit 1; fi
}
run_fail() {
  echo ">>> FAIL (expected): $*"
  set +e; set -x; "$@"; rc=$?; set +x; set -e
  if [ $rc -eq 0 ]; then echo "Expected failure but succeeded: $*" >&2; exit 1; fi
}

# Basic sanity
run_ok go run ./cmd/flight-booking version
run_ok go run ./cmd/flight-booking db:ping

echo "[5a/6] Airport scenarios"
# Initial list (seeded CGK/DPS)
run_ok go run ./cmd/flight-booking airport list --limit 10
# Create a few airports
run_ok go run ./cmd/flight-booking airport create --code SUB --city "Surabaya"
run_ok go run ./cmd/flight-booking airport create --code UPG --city "Makassar"
# Duplicate should fail
run_fail go run ./cmd/flight-booking airport create --code SUB --city "Surabaya"
# Pagination sample
run_ok go run ./cmd/flight-booking airport list --limit 1 --offset 1
# Update
run_ok go run ./cmd/flight-booking airport update --code UPG --city "Makassar (Ujung Pandang)"
# Delete non-existent (expect failure)
run_fail go run ./cmd/flight-booking airport delete XYZ
# Delete and verify list
run_ok go run ./cmd/flight-booking airport delete SUB
run_ok go run ./cmd/flight-booking airport delete UPG
run_ok go run ./cmd/flight-booking airport list --limit 10

echo "[5b/6] Airplane scenarios"
# Create airplanes
run_ok go run ./cmd/flight-booking airplane create --code B737 --seats 180
run_ok go run ./cmd/flight-booking airplane create --code A320 --seats 150
# Duplicate and invalid seats
run_fail go run ./cmd/flight-booking airplane create --code B737 --seats 180
run_fail go run ./cmd/flight-booking airplane create --code INV --seats 0
# Pagination and list
run_ok go run ./cmd/flight-booking airplane list --limit 1 --offset 1
# Update seats and non-existent update
run_ok go run ./cmd/flight-booking airplane update --code B737 --seats 200
run_fail go run ./cmd/flight-booking airplane update --code NONE --seats 100
# Delete and non-existent delete
run_ok go run ./cmd/flight-booking airplane delete B737
run_ok go run ./cmd/flight-booking airplane delete A320
run_fail go run ./cmd/flight-booking airplane delete NONE
run_ok go run ./cmd/flight-booking airplane list --limit 10

echo "[6/6] Done. Container will be removed."
