//go:build e2e

package e2e

import (
    "bytes"
    "context"
    "fmt"
    "os"
    "path/filepath"
    "runtime"
    "strings"
    "testing"
    "time"

    _ "github.com/jackc/pgx/v5/stdlib"
    "github.com/jmoiron/sqlx"
    "github.com/pressly/goose/v3"
    tc "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"

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
        WaitingFor: wait.ForAll(
            wait.ForListeningPort("5432/tcp"),
            wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
        ).WithStartupTimeout(60 * time.Second),
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
    // Base connection as postgres
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
    // Run migrations as migrator role so ownership/default privileges apply
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

