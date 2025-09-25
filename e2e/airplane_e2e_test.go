//go:build e2e

package e2e

import (
    "os"
    "strings"
    "testing"
)

func TestAirplaneE2E(t *testing.T) {
    dsn, terminate := startPostgres(t)
    defer terminate()
    applyBootstrap(t, dsn)

    // Configure CLI env to use app role
    parts := strings.Split(strings.TrimPrefix(dsn, "postgres://"), "@")
    hostportDB := parts[1]
    hp, dbpart := strings.Split(hostportDB, "/")[0], strings.Split(hostportDB, "/")[1]
    host := strings.Split(hp, ":")[0]
    port := strings.Split(hp, ":")[1]
    dbname := strings.Split(dbpart, "?")[0]
    os.Setenv("FLIGHT_DB_HOST", host)
    os.Setenv("FLIGHT_DB_PORT", port)
    os.Setenv("FLIGHT_DB_USER", "flight_app")
    os.Setenv("FLIGHT_DB_PASSWORD", "app")
    os.Setenv("FLIGHT_DB_NAME", dbname)
    os.Setenv("FLIGHT_DB_SSLMODE", "disable")

    // Create airplanes
    if _, err := runCLI("airplane", "create", "--code", "B737", "--seats", "180"); err != nil {
        t.Fatalf("create B737: %v", err)
    }
    if _, err := runCLI("airplane", "create", "--code", "A320", "--seats", "150"); err != nil {
        t.Fatalf("create A320: %v", err)
    }
    // Duplicate and invalid seats
    if _, err := runCLI("airplane", "create", "--code", "B737", "--seats", "180"); err == nil {
        t.Fatalf("expected duplicate airplane error")
    }
    if _, err := runCLI("airplane", "create", "--code", "INV", "--seats", "0"); err == nil {
        t.Fatalf("expected invalid seats error")
    }
    // Pagination and list
    if _, err := runCLI("airplane", "list", "--limit", "1", "--offset", "1"); err != nil {
        t.Fatalf("airplane list pagination: %v", err)
    }
    // Update seats and non-existent update
    if _, err := runCLI("airplane", "update", "--code", "B737", "--seats", "200"); err != nil {
        t.Fatalf("update B737 seats: %v", err)
    }
    if _, err := runCLI("airplane", "update", "--code", "NONE", "--seats", "100"); err == nil {
        t.Fatalf("expected update non-existent airplane error")
    }
    // Delete and non-existent delete
    if _, err := runCLI("airplane", "delete", "B737"); err != nil { t.Fatalf("delete B737: %v", err) }
    if _, err := runCLI("airplane", "delete", "A320"); err != nil { t.Fatalf("delete A320: %v", err) }
    if _, err := runCLI("airplane", "delete", "NONE"); err == nil { t.Fatalf("expected delete non-existent airplane error") }
}

