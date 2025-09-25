package sqlxrepo

import (
    "context"
    "fmt"
    "regexp"
    "testing"
    "time"

    "github.com/DATA-DOG/go-sqlmock"
    "github.com/ambiyansyah-risyal/flight-booking/internal/domain"
    "github.com/jmoiron/sqlx"
)

func newMockAirDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock, func()) {
    t.Helper()
    db, mock, err := sqlmock.New()
    if err != nil { t.Fatalf("sqlmock: %v", err) }
    return sqlx.NewDb(db, "pgx"), mock, func(){ _ = db.Close() }
}

func TestAirplaneRepo_Create_List_Update_Delete(t *testing.T) {
    db, mock, cleanup := newMockAirDB(t)
    defer cleanup()
    repo := NewAirplaneRepository(db)
    now := time.Now()
    mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO airplanes (code, seat_capacity) VALUES ($1,$2) RETURNING id, created_at`)).
        WithArgs("B737", 180).WillReturnRows(sqlmock.NewRows([]string{"id","created_at"}).AddRow(1, now))
    a := &domain.Airplane{Code:"B737", SeatCapacity:180}
    if err := repo.Create(context.Background(), a); err != nil { t.Fatalf("create: %v", err) }

    mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, code, seat_capacity, created_at FROM airplanes ORDER BY code LIMIT $1 OFFSET $2`)).
        WithArgs(10, 0).
        WillReturnRows(sqlmock.NewRows([]string{"id","code","seat_capacity","created_at"}).AddRow(1,"B737",180, now))
    items, err := repo.List(context.Background(), 10, 0)
    if err != nil || len(items) != 1 { t.Fatalf("list: %v n=%d", err, len(items)) }

    mock.ExpectExec(regexp.QuoteMeta(`UPDATE airplanes SET seat_capacity=$2 WHERE code=$1`)).
        WithArgs("B737", 200).WillReturnResult(sqlmock.NewResult(0,1))
    if err := repo.UpdateSeats(context.Background(), "B737", 200); err != nil { t.Fatalf("update: %v", err) }

    mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM airplanes WHERE code=$1`)).
        WithArgs("B737").WillReturnResult(sqlmock.NewResult(0,1))
    if err := repo.Delete(context.Background(), "B737"); err != nil { t.Fatalf("delete: %v", err) }
}

type pqErr struct{ msg string }
func (e *pqErr) Error() string { return e.msg }

func TestAirplaneRepo_Duplicate_NotFound_Paths(t *testing.T) {
    db, mock, cleanup := newMockAirDB(t)
    defer cleanup()
    repo := NewAirplaneRepository(db)

    mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO airplanes (code, seat_capacity) VALUES ($1,$2) RETURNING id, created_at`)).
        WithArgs("B737", 180).WillReturnError(&pqErr{msg:"duplicate key value violates unique constraint"})
    if err := repo.Create(context.Background(), &domain.Airplane{Code:"B737", SeatCapacity:180}); err != domain.ErrAirplaneExists {
        t.Fatalf("want exists, got %v", err)
    }

    mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, code, seat_capacity, created_at FROM airplanes WHERE code=$1`)).
        WithArgs("NONE").WillReturnRows(sqlmock.NewRows([]string{"id","code","seat_capacity","created_at"}))
    if _, err := repo.GetByCode(context.Background(), "NONE"); err != domain.ErrAirplaneNotFound {
        t.Fatalf("want not found, got %v", err)
    }

    mock.ExpectExec(regexp.QuoteMeta(`UPDATE airplanes SET seat_capacity=$2 WHERE code=$1`)).
        WithArgs("NONE", 100).WillReturnResult(sqlmock.NewResult(0,0))
    if err := repo.UpdateSeats(context.Background(), "NONE", 100); err != domain.ErrAirplaneNotFound {
        t.Fatalf("want not found update, got %v", err)
    }

    mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM airplanes WHERE code=$1`)).
        WithArgs("NONE").WillReturnResult(sqlmock.NewResult(0,0))
    if err := repo.Delete(context.Background(), "NONE"); err != domain.ErrAirplaneNotFound {
        t.Fatalf("want not found delete, got %v", err)
    }
}

func TestAirplaneRepo_ListErrors_And_GetByCode(t *testing.T) {
    db, mock, cleanup := newMockAirDB(t)
    defer cleanup()
    repo := NewAirplaneRepository(db)
    // query error
    mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, code, seat_capacity, created_at FROM airplanes ORDER BY code LIMIT $1 OFFSET $2`)).
        WithArgs(5, 0).WillReturnError(fmt.Errorf("db down"))
    if _, err := repo.List(context.Background(), 5, 0); err == nil { t.Fatalf("expected error") }

    // rows error
    now := time.Now()
    rows := sqlmock.NewRows([]string{"id","code","seat_capacity","created_at"}).AddRow(1, "A320", 150, now)
    rows.RowError(0, fmt.Errorf("scan error"))
    mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, code, seat_capacity, created_at FROM airplanes ORDER BY code LIMIT $1 OFFSET $2`)).
        WithArgs(5, 0).WillReturnRows(rows)
    if _, err := repo.List(context.Background(), 5, 0); err == nil { t.Fatalf("expected rows error") }

    // get by code success
    mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, code, seat_capacity, created_at FROM airplanes WHERE code=$1`)).
        WithArgs("A320").WillReturnRows(sqlmock.NewRows([]string{"id","code","seat_capacity","created_at"}).AddRow(2, "A320", 150, now))
    a, err := repo.GetByCode(context.Background(), "A320")
    if err != nil || a.Code != "A320" { t.Fatalf("get success err=%v a=%+v", err, a) }
}

func TestAirplaneRepo_UpdateDelete_ErrExec(t *testing.T) {
    db, mock, cleanup := newMockAirDB(t)
    defer cleanup()
    repo := NewAirplaneRepository(db)
    mock.ExpectExec(regexp.QuoteMeta(`UPDATE airplanes SET seat_capacity=$2 WHERE code=$1`)).
        WithArgs("B737", 123).WillReturnError(fmt.Errorf("exec fail"))
    if err := repo.UpdateSeats(context.Background(), "B737", 123); err == nil { t.Fatalf("expected exec error") }

    mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM airplanes WHERE code=$1`)).
        WithArgs("B737").WillReturnError(fmt.Errorf("exec fail"))
    if err := repo.Delete(context.Background(), "B737"); err == nil { t.Fatalf("expected exec error") }
}
