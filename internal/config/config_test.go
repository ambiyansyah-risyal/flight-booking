package config

import (
    "os"
    "testing"
)

func TestDefaultsAndDSN(t *testing.T) {
    t.Setenv("FLIGHT_DB_HOST", "localhost")
    t.Setenv("FLIGHT_DB_PORT", "5432")
    t.Setenv("FLIGHT_DB_USER", "u")
    t.Setenv("FLIGHT_DB_PASSWORD", "p@ss")
    t.Setenv("FLIGHT_DB_NAME", "flight")
    t.Setenv("FLIGHT_DB_SSLMODE", "disable")

    cfg, err := Load()
    if err != nil { t.Fatalf("load: %v", err) }
    dsn := cfg.Database.DSN()
    want := "postgres://u:p%40ss@localhost:5432/flight?sslmode=disable"
    if dsn != want {
        t.Fatalf("dsn mismatch: got %q want %q", dsn, want)
    }
}

func TestEnvFallbackWithoutPrefix(t *testing.T) {
    // When plain env exists, it should also be considered
    os.Unsetenv("FLIGHT_DB_HOST")
    t.Setenv("DB_HOST", "127.0.0.1")
    cfg, err := Load()
    if err != nil { t.Fatalf("load: %v", err) }
    if cfg.Database.Host != "127.0.0.1" {
        t.Fatalf("expected host from DB_HOST, got %q", cfg.Database.Host)
    }
}

