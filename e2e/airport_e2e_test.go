//go:build e2e

package e2e

import (
    "bytes"
    "context"
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "testing"
    "time"

    _ "github.com/jackc/pgx/v5/stdlib"
    "github.com/jmoiron/sqlx"
    "github.com/pressly/goose/v3"
    tc "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
    "runtime"

    "github.com/ambiyansyah-risyal/flight-booking/internal/adapter/cli"
)

func startPostgres(t *testing.T) (dsn string, terminate func()) {
    t.Helper()
    ctx := context.Background()
    req := tc.ContainerRequest{
        Image:        "postgres:16-alpine",
        ExposedPorts: []string{"5432/tcp"},
        Env: map[string]string{
            "POSTGRES_PASSWORD": "postgres",
            "POSTGRES_DB":       "flight",
            "POSTGRES_USER":     "postgres",
        },
        WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
    }
    container, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{ContainerRequest: req, Started: true})
    if err != nil { t.Fatalf("start container: %v", err) }
    host, err := container.Host(ctx)
    if err != nil { t.Fatalf("host: %v", err) }
    port, err := container.MappedPort(ctx, "5432/tcp")
    if err != nil { t.Fatalf("mapped port: %v", err) }
    dsn = fmt.Sprintf("postgres://postgres:postgres@%s:%s/flight?sslmode=disable", host, port.Port())
    term := func() { _ = container.Terminate(ctx) }
    return dsn, term
}

func applyBootstrap(t *testing.T, dsn string) {
    t.Helper()
    db, err := sqlx.Open("pgx", dsn)
    if err != nil { t.Fatalf("open db: %v", err) }
    defer db.Close()
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    stmts := []string{
        `DO $$
        BEGIN
           IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'flight_migrator') THEN
              CREATE ROLE flight_migrator LOGIN PASSWORD 'migrator';
           END IF;
           IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'flight_app') THEN
              CREATE ROLE flight_app LOGIN PASSWORD 'app';
           END IF;
        END$$;`,
        `GRANT ALL PRIVILEGES ON DATABASE flight TO flight_migrator;`,
        `GRANT USAGE, CREATE ON SCHEMA public TO flight_migrator;`,
        `GRANT CONNECT ON DATABASE flight TO flight_app;`,
        `GRANT USAGE ON SCHEMA public TO flight_app;`,
        `ALTER DEFAULT PRIVILEGES FOR ROLE flight_migrator IN SCHEMA public
            GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO flight_app;`,
        `ALTER DEFAULT PRIVILEGES FOR ROLE flight_migrator IN SCHEMA public
            GRANT USAGE, SELECT ON SEQUENCES TO flight_app;`,
    }
    for _, s := range stmts {
        if _, err := db.ExecContext(ctx, s); err != nil {
            t.Fatalf("bootstrap exec failed: %v\nSQL: %s", err, s)
        }
    }
    if err := goose.SetDialect("postgres"); err != nil { t.Fatalf("dialect: %v", err) }
    // Run migrations as migrator role so ownership and default privs apply
    parts := strings.Split(strings.TrimPrefix(dsn, "postgres://"), "@")
    hpdb := parts[1]
    hp := strings.Split(hpdb, "/")[0]
    migratorDSN := fmt.Sprintf("postgres://flight_migrator:migrator@%s/flight?sslmode=disable", hp)
    dbm, err := sqlx.Open("pgx", migratorDSN)
    if err != nil { t.Fatalf("open migrator db: %v", err) }
    defer dbm.Close()
    // Locate migrations dir relative to repo root
    _, thisFile, _, _ := runtime.Caller(0)
    migrationsDir := filepath.Join(filepath.Dir(thisFile), "..", "migrations")
    if err := goose.Up(dbm.DB, migrationsDir); err != nil { t.Fatalf("migrate up: %v", err) }
}

func captureOutput(fn func() error) (string, error) {
    old := os.Stdout
    r, w, _ := os.Pipe()
    os.Stdout = w
    err := fn()
    _ = w.Close()
    os.Stdout = old
    var buf bytes.Buffer
    _, _ = buf.ReadFrom(r)
    return buf.String(), err
}

func runCLI(args ...string) (string, error) {
    os.Args = append([]string{"flight-booking"}, args...)
    return captureOutput(func() error { return cli.Execute() })
}

func TestAirportE2E(t *testing.T) {
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

    // 1) Create new airports (seeded CGK/DPS exist already)
    if _, err := runCLI("airport", "create", "--code", "SUB", "--city", "Surabaya"); err != nil {
        t.Fatalf("create SUB: %v", err)
    }
    if _, err := runCLI("airport", "create", "--code", "UPG", "--city", "Makassar"); err != nil {
        t.Fatalf("create UPG: %v", err)
    }
    // Duplicate create should error
    if _, err := runCLI("airport", "create", "--code", "SUB", "--city", "Surabaya"); err == nil {
        t.Fatalf("expected duplicate create error")
    }

    // 2) List with pagination
    out, err := runCLI("airport", "list", "--limit", "3", "--offset", "0")
    if err != nil { t.Fatalf("list: %v", err) }
    if !(strings.Contains(out, "SUB") && strings.Contains(out, "Surabaya")) { t.Fatalf("unexpected list content: %s", out) }

    // 3) Update UPG city name and verify via list
    if _, err := runCLI("airport", "update", "--code", "UPG", "--city", "Makassar (Ujung Pandang)"); err != nil {
        t.Fatalf("update UPG: %v", err)
    }
    out, err = runCLI("airport", "list", "--limit", "10")
    if err != nil { t.Fatalf("list after update: %v", err) }
    if !(strings.Contains(out, "UPG") && strings.Contains(out, "Makassar (Ujung Pandang)")) { t.Fatalf("updated city not found: %s", out) }

    // 4) Delete DPS (seeded) and ensure it disappears
    if _, err := runCLI("airport", "delete", "DPS"); err != nil { t.Fatalf("delete DPS: %v", err) }
    out, err = runCLI("airport", "list", "--limit", "50")
    if err != nil { t.Fatalf("list after delete: %v", err) }
    if strings.Contains(out, "DPS") || strings.Contains(out, "Denpasar") { t.Fatalf("DPS still present after delete: %s", out) }
}
