package cli

import (
    "context"
    "os"
    "testing"

    "github.com/ambiyansyah-risyal/flight-booking/internal/domain"
    "github.com/jmoiron/sqlx"
    "github.com/DATA-DOG/go-sqlmock"
    "fmt"
)

type fakePlaneRepo struct{ data map[string]int }
func (f *fakePlaneRepo) Create(ctx context.Context, a *domain.Airplane) error { f.data[a.Code]=a.SeatCapacity; return nil }
func (f *fakePlaneRepo) GetByCode(ctx context.Context, code string) (*domain.Airplane, error) { if s,ok:=f.data[code]; ok { return &domain.Airplane{Code:code, SeatCapacity:s}, nil }; return nil, domain.ErrAirplaneNotFound }
func (f *fakePlaneRepo) List(ctx context.Context, limit, offset int) ([]domain.Airplane, error) { out:=[]domain.Airplane{}; for k,v:=range f.data { out=append(out, domain.Airplane{Code:k, SeatCapacity:v})}; return out, nil }
func (f *fakePlaneRepo) UpdateSeats(ctx context.Context, code string, seats int) error { if _,ok:=f.data[code]; !ok { return domain.ErrAirplaneNotFound }; f.data[code]=seats; return nil }
func (f *fakePlaneRepo) Delete(ctx context.Context, code string) error { if _,ok:=f.data[code]; !ok { return domain.ErrAirplaneNotFound }; delete(f.data, code); return nil }

func TestAirplaneCLI_Flow(t *testing.T) {
    oldDB, oldRepo := newAirplaneDB, newAirplaneRepoF
    t.Cleanup(func(){ newAirplaneDB=oldDB; newAirplaneRepoF=oldRepo })
    newAirplaneDB = func(dsn string) (*sqlx.DB, error) {
        db, _, _ := sqlmock.New()
        return sqlx.NewDb(db, "pgx"), nil
    }
    r := &fakePlaneRepo{data: map[string]int{"B737":180}}
    newAirplaneRepoF = func(db *sqlx.DB) domain.AirplaneRepository { return r }

    t.Setenv("FLIGHT_DB_HOST", "localhost")

    os.Args = []string{"flight-booking", "airplane", "create", "--code", "A320", "--seats", "150"}
    if err := Execute(); err != nil { t.Fatalf("create: %v", err) }

    os.Args = []string{"flight-booking", "airplane", "update", "--code", "A320", "--seats", "160"}
    if err := Execute(); err != nil { t.Fatalf("update: %v", err) }

    os.Args = []string{"flight-booking", "airplane", "list"}
    if err := Execute(); err != nil { t.Fatalf("list: %v", err) }

    os.Args = []string{"flight-booking", "airplane", "delete", "A320"}
    if err := Execute(); err != nil { t.Fatalf("delete: %v", err) }
}

func TestAirplaneCLI_OpenError(t *testing.T) {
    oldDB := newAirplaneDB
    t.Cleanup(func(){ newAirplaneDB = oldDB })
    newAirplaneDB = func(dsn string) (*sqlx.DB, error) { return nil, fmt.Errorf("open error") }
    t.Setenv("FLIGHT_DB_HOST", "localhost")
    os.Args = []string{"flight-booking", "airplane", "list"}
    if err := Execute(); err == nil { t.Fatalf("expected open error") }
}

func TestAirplaneCLI_MissingFlags(t *testing.T) {
    t.Setenv("FLIGHT_DB_HOST", "localhost")
    os.Args = []string{"flight-booking", "airplane", "create"}
    if err := Execute(); err == nil { t.Fatalf("expected flag error") }
}

func TestAirplaneCLI_UpdateMissingFlags(t *testing.T) {
    t.Setenv("FLIGHT_DB_HOST", "localhost")
    os.Args = []string{"flight-booking", "airplane", "update"}
    if err := Execute(); err == nil { t.Fatalf("expected flag error") }
}

func TestAirplaneCLI_DeleteNotFound(t *testing.T) {
    oldDB, oldRepo := newAirplaneDB, newAirplaneRepoF
    t.Cleanup(func(){ newAirplaneDB=oldDB; newAirplaneRepoF=oldRepo })
    newAirplaneDB = func(dsn string) (*sqlx.DB, error) { db,_,_ := sqlmock.New(); return sqlx.NewDb(db, "pgx"), nil }
    r := &fakePlaneRepo{data: map[string]int{"EXIST":100}}
    newAirplaneRepoF = func(db *sqlx.DB) domain.AirplaneRepository { return r }
    t.Setenv("FLIGHT_DB_HOST", "localhost")
    os.Args = []string{"flight-booking", "airplane", "delete", "NONE"}
    if err := Execute(); err == nil { t.Fatalf("expected not found error") }
}
