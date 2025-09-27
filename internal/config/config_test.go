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
    if cfg.Database.MaxOpenConns == 0 || cfg.Database.ConnMaxLife == 0 {
        t.Fatalf("expected defaults for pool settings")
    }
}

func TestEnvFallbackWithoutPrefix(t *testing.T) {
    // When plain env exists, it should also be considered
    _ = os.Unsetenv("FLIGHT_DB_HOST")
    t.Setenv("DB_HOST", "127.0.0.1")
    cfg, err := Load()
    if err != nil { t.Fatalf("load: %v", err) }
    if cfg.Database.Host != "127.0.0.1" {
        t.Fatalf("expected host from DB_HOST, got %q", cfg.Database.Host)
    }
}

func TestDBURLOverride(t *testing.T) {
    t.Setenv("FLIGHT_DATABASE_URL", "postgres://user:pass@db:5432/name?sslmode=require")
    cfg, err := Load()
    if err != nil { t.Fatalf("load: %v", err) }
    if cfg.Database.DSN() != "postgres://user:pass@db:5432/name?sslmode=require" {
        t.Fatalf("expected URL override, got %s", cfg.Database.DSN())
    }
}

func TestInvalidSSLMode(t *testing.T) {
    t.Setenv("FLIGHT_DB_SSLMODE", "weird")
    if _, err := Load(); err == nil {
        t.Fatalf("expected sslmode validation error")
    }
}

func TestInvalidPort(t *testing.T) {
    t.Setenv("FLIGHT_DB_HOST", "localhost")
    t.Setenv("FLIGHT_DB_PORT", "0")
    if _, err := Load(); err == nil {
        t.Fatalf("expected port validation error")
    }
}

func TestInvalidPortTooHigh(t *testing.T) {
    t.Setenv("FLIGHT_DB_HOST", "localhost")
    t.Setenv("FLIGHT_DB_PORT", "65536")  // Above valid range
    if _, err := Load(); err == nil {
        t.Fatalf("expected port validation error for port too high")
    }
}

func TestEmptyHost(t *testing.T) {
    // To test empty host validation, we need to override the default value
    // Since the default is "localhost", we need to ensure it can still be empty
    t.Setenv("FLIGHT_DB_HOST", "localhost") // Start with valid host
    cfg, err := Load()
    if err != nil {
        t.Fatalf("unexpected error with valid host: %v", err)
    }
    
    // Now manually test the validation function with an empty host
    cfg.Database.Host = ""  // Set host to empty after loading
    if err := validate(cfg); err == nil {
        t.Fatalf("expected validation error for empty host")
    }
}

func TestWhitespaceHost(t *testing.T) {
    t.Setenv("FLIGHT_DB_HOST", "localhost") // Start with valid host
    cfg, err := Load()
    if err != nil {
        t.Fatalf("unexpected error with valid host: %v", err)
    }
    
    // Now manually test the validation function with a whitespace-only host
    cfg.Database.Host = "   "  // Set host to whitespace
    if err := validate(cfg); err == nil {
        t.Fatalf("expected validation error for whitespace-only host")
    }
}

func TestDiscoverConfig_NoFile(t *testing.T) {
    got := DiscoverConfig()
    if len(got) != 0 {
        t.Fatalf("expected no files discovered, got %v", got)
    }
}
