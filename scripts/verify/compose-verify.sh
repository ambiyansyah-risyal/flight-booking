#!/usr/bin/env bash
set -euo pipefail

# Compose-based verification script.
# - Builds images for app and migrate
# - Starts DB, waits for readiness
# - Runs Goose migrations with migrate service
# - Exercises CLI flows inside the app container for airport and airplane

COMPOSE_FILE=${COMPOSE_FILE:-docker/compose.yml}

cleanup() {
  docker compose -f "$COMPOSE_FILE" down -v >/dev/null 2>&1 || true
}
trap cleanup EXIT

echo "[1/5] Building images (app, migrate)"
docker compose -f "$COMPOSE_FILE" build app migrate

echo "[2/5] Starting DB and waiting for readiness"
docker compose -f "$COMPOSE_FILE" up -d db
echo -n "Waiting for Postgres (compose)"
for i in {1..60}; do
  if docker compose -f "$COMPOSE_FILE" exec -T db pg_isready -U postgres -d flight >/dev/null 2>&1; then break; fi
  printf "."; sleep 1
  if [ "$i" -eq 60 ]; then echo "\nERROR: DB not ready" >&2; exit 1; fi
done
echo

echo "[3/5] Running migrations (migrate service)"
docker compose -f "$COMPOSE_FILE" run --rm migrate up

# Helpers
run_ok() {
  echo ">>> OK: $*"
  set -x; docker compose -f "$COMPOSE_FILE" run --rm app "$@"; rc=$?; set +x
  if [ $rc -ne 0 ]; then echo "Expected OK but failed: $*" >&2; exit 1; fi
}
run_fail() {
  echo ">>> FAIL (expected): $*"
  set +e; set -x; docker compose -f "$COMPOSE_FILE" run --rm app "$@"; rc=$?; set +x; set -e
  if [ $rc -eq 0 ]; then echo "Expected failure but succeeded: $*" >&2; exit 1; fi
}

echo "[4/5] Sanity checks"
run_ok version
run_ok db:ping

echo "[4a/5] Airport scenarios"
run_ok airport list --limit 10
run_ok airport create --code SUB --city Surabaya
run_ok airport create --code UPG --city Makassar
run_fail airport create --code SUB --city Surabaya
run_ok airport list --limit 1 --offset 1
run_ok airport update --code UPG --city "Makassar (Ujung Pandang)"
run_fail airport delete XYZ
run_ok airport delete SUB
run_ok airport delete UPG
run_ok airport list --limit 10

echo "[4b/5] Airplane scenarios"
run_ok airplane create --code B737 --seats 180
run_ok airplane create --code A320 --seats 150
run_fail airplane create --code B737 --seats 180
run_fail airplane create --code INV --seats 0
run_ok airplane list --limit 1 --offset 1
run_ok airplane update --code B737 --seats 200
run_fail airplane update --code NONE --seats 100
run_ok airplane delete B737
run_ok airplane delete A320
run_fail airplane delete NONE
run_ok airplane list --limit 10

echo "[5/5] Done (stack will be torn down)"

