package cli

import (
    "bytes"
    "context"
    "fmt"
    "os"
    "testing"

    "github.com/DATA-DOG/go-sqlmock"
    "github.com/ambiyansyah-risyal/flight-booking/internal/domain"
    "github.com/jmoiron/sqlx"
    "strings"
)

func captureOutput(fn func()) string {
    old := os.Stdout
    r, w, _ := os.Pipe()
    os.Stdout = w
    fn()
    _ = w.Close()
    os.Stdout = old
    var buf bytes.Buffer
    _, _ = buf.ReadFrom(r)
    return buf.String()
}

func TestExecute_Version(t *testing.T) {
    t.Setenv("FLIGHT_DB_HOST", "localhost")
    os.Args = []string{"flight-booking", "version"}
    out := captureOutput(func() {
        if err := Execute(); err != nil { t.Fatalf("execute: %v", err) }
    })
    if out == "" { t.Fatalf("expected version output") }
}

func TestExecute_Help(t *testing.T) {
    t.Setenv("FLIGHT_DB_HOST", "localhost")
    os.Args = []string{"flight-booking"}
    out := captureOutput(func() {
        if err := Execute(); err != nil { t.Fatalf("execute: %v", err) }
    })
    if out == "" { t.Fatalf("expected help output") }
}

func TestVersion_WithBuildMeta(t *testing.T) {
    // Set build metadata to exercise version output branch
    oldV, oldC, oldD := Version, Commit, BuildDate
    Version, Commit, BuildDate = "1.2.3", "abc1234", "2025-01-01T00:00:00Z"
    t.Cleanup(func(){ Version, Commit, BuildDate = oldV, oldC, oldD })

    t.Setenv("FLIGHT_DB_HOST", "localhost")
    os.Args = []string{"flight-booking", "version"}
    out := captureOutput(func(){ _ = Execute() })
    if !(strings.Contains(out, "1.2.3") && strings.Contains(out, "abc1234") && strings.Contains(out, "2025-01-01")) {
        t.Fatalf("unexpected version output: %s", out)
    }
}

// Fake repo for CLI tests
type fakeRepo struct{ data map[string]string }
func (f *fakeRepo) Create(ctx context.Context, a *domain.Airport) error { f.data[a.Code]=a.City; return nil }
func (f *fakeRepo) GetByCode(ctx context.Context, code string) (*domain.Airport, error) { if c,ok:=f.data[code]; ok { return &domain.Airport{Code:code, City:c}, nil }; return nil, domain.ErrAirportNotFound }
func (f *fakeRepo) List(ctx context.Context, limit, offset int) ([]domain.Airport, error) { out:=[]domain.Airport{}; for k,v:= range f.data { out=append(out, domain.Airport{Code:k, City:v}) }; return out, nil }
func (f *fakeRepo) Update(ctx context.Context, code string, city string) error { if _,ok:=f.data[code]; !ok { return domain.ErrAirportNotFound }; f.data[code]=city; return nil }
func (f *fakeRepo) Delete(ctx context.Context, code string) error { if _,ok:=f.data[code]; !ok { return domain.ErrAirportNotFound }; delete(f.data, code); return nil }

func TestAirportCLI_Subcommands(t *testing.T) {
    // Patch DB and repo factories
    oldNewDB := newDB
    oldNewRepo := newAirportRepo
    t.Cleanup(func(){ newDB = oldNewDB; newAirportRepo = oldNewRepo })

    mockDB, _, err := sqlmock.New()
    if err != nil { t.Fatalf("sqlmock: %v", err) }
    sqldb := sqlx.NewDb(mockDB, "pgx")
    newDB = func(dsn string) (*sqlx.DB, error) { return sqldb, nil }
    r := &fakeRepo{data: map[string]string{"CGK":"Jakarta"}}
    newAirportRepo = func(db *sqlx.DB) domain.AirportRepository { return r }

    t.Setenv("FLIGHT_DB_HOST", "localhost")

    // Create
    os.Args = []string{"flight-booking", "airport", "create", "--code", "DPS", "--city", "Denpasar"}
    if err := Execute(); err != nil { t.Fatalf("create: %v", err) }
    // Update
    os.Args = []string{"flight-booking", "airport", "update", "--code", "DPS", "--city", "Bali"}
    if err := Execute(); err != nil { t.Fatalf("update: %v", err) }
    // List
    os.Args = []string{"flight-booking", "airport", "list"}
    out := captureOutput(func(){ _ = Execute() })
    if out == "" { t.Fatalf("expected list output") }
    // Delete
    os.Args = []string{"flight-booking", "airport", "delete", "DPS"}
    if err := Execute(); err != nil { t.Fatalf("delete: %v", err) }
}

func TestAirportCLI_DBInitError(t *testing.T) {
    oldNewDB := newDB
    t.Cleanup(func(){ newDB = oldNewDB })
    newDB = func(dsn string) (*sqlx.DB, error) { return nil, fmt.Errorf("open error") }
    t.Setenv("FLIGHT_DB_HOST", "localhost")
    os.Args = []string{"flight-booking", "airport", "list"}
    if err := Execute(); err == nil { t.Fatalf("expected error from db open") }
}

func TestAirportCLI_CreateMissingFlags(t *testing.T) {
    t.Setenv("FLIGHT_DB_HOST", "localhost")
    os.Args = []string{"flight-booking", "airport", "create"}
    if err := Execute(); err == nil { t.Fatalf("expected error due to required flags") }
}

func TestAirportCLI_UpdateMissingFlags(t *testing.T) {
    t.Setenv("FLIGHT_DB_HOST", "localhost")
    os.Args = []string{"flight-booking", "airport", "update"}
    if err := Execute(); err == nil { t.Fatalf("expected error due to required flags") }
}

func TestAirportCLI_DeleteNotFound(t *testing.T) {
    oldNewDB := newDB
    oldNewRepo := newAirportRepo
    t.Cleanup(func(){ newDB = oldNewDB; newAirportRepo = oldNewRepo })
    mockDB, _, err := sqlmock.New()
    if err != nil { t.Fatalf("sqlmock: %v", err) }
    sqldb := sqlx.NewDb(mockDB, "pgx")
    newDB = func(dsn string) (*sqlx.DB, error) { return sqldb, nil }
    r := &fakeRepo{data: map[string]string{"CGK":"Jakarta"}}
    newAirportRepo = func(db *sqlx.DB) domain.AirportRepository { return r }

    t.Setenv("FLIGHT_DB_HOST", "localhost")
    os.Args = []string{"flight-booking", "airport", "delete", "XXX"}
    if err := Execute(); err == nil { t.Fatalf("expected not found error") }
}

func TestDBPing_FailFast(t *testing.T) {
    // Use an unreachable port to cause quick failure
    t.Setenv("FLIGHT_DB_HOST", "127.0.0.1")
    t.Setenv("FLIGHT_DB_PORT", "1")
    t.Setenv("FLIGHT_DB_USER", "u")
    t.Setenv("FLIGHT_DB_PASSWORD", "p")
    t.Setenv("FLIGHT_DB_NAME", "db")
    t.Setenv("FLIGHT_DB_SSLMODE", "disable")
    os.Args = []string{"flight-booking", "db:ping"}
    if err := Execute(); err == nil { t.Fatalf("expected ping error") }
}
