package sqlxrepo

import (
    "context"
    "regexp"
    "testing"
    "time"

    "github.com/DATA-DOG/go-sqlmock"
    "github.com/ambiyansyah-risyal/flight-booking/internal/domain"
    "github.com/jmoiron/sqlx"
)

func newMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock, func()) {
    t.Helper()
    db, mock, err := sqlmock.New()
    if err != nil { t.Fatalf("sqlmock: %v", err) }
    return sqlx.NewDb(db, "pgx"), mock, func() { db.Close() }
}

func TestAirportRepo_Create_Success(t *testing.T) {
    db, mock, cleanup := newMockDB(t)
    defer cleanup()
    repo := NewAirportRepository(db)
    createdAt := time.Now()
    mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO airports (code, city) VALUES ($1, $2) RETURNING id, created_at`)).
        WithArgs("CGK", "Jakarta").
        WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(1, createdAt))

    a := &domain.Airport{Code: "CGK", City: "Jakarta"}
    if err := repo.Create(context.Background(), a); err != nil { t.Fatalf("create: %v", err) }
    if a.ID != 1 || a.CreatedAt == "" { t.Fatalf("unexpected airport: %+v", a) }
}

func TestAirportRepo_Create_Duplicate(t *testing.T) {
    db, mock, cleanup := newMockDB(t)
    defer cleanup()
    repo := NewAirportRepository(db)
    mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO airports (code, city) VALUES ($1, $2) RETURNING id, created_at`)).
        WithArgs("CGK", "Jakarta").
        WillReturnError(&pqError{msg: "duplicate key value violates unique constraint \"airports_code_key\""})
    a := &domain.Airport{Code: "CGK", City: "Jakarta"}
    if err := repo.Create(context.Background(), a); err != domain.ErrAirportExists {
        t.Fatalf("want ErrAirportExists, got %v", err)
    }
}

func TestAirportRepo_GetByCode_NotFound(t *testing.T) {
    db, mock, cleanup := newMockDB(t)
    defer cleanup()
    repo := NewAirportRepository(db)
    mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, code, city, created_at FROM airports WHERE code=$1`)).
        WithArgs("XXX").
        WillReturnRows(sqlmock.NewRows([]string{"id", "code", "city", "created_at"}))
    if _, err := repo.GetByCode(context.Background(), "XXX"); err != domain.ErrAirportNotFound {
        t.Fatalf("want not found, got %v", err)
    }
}

func TestAirportRepo_List_Update_Delete(t *testing.T) {
    db, mock, cleanup := newMockDB(t)
    defer cleanup()
    repo := NewAirportRepository(db)
    now := time.Now()
    mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, code, city, created_at FROM airports ORDER BY code LIMIT $1 OFFSET $2`)).
        WithArgs(2, 0).
        WillReturnRows(sqlmock.NewRows([]string{"id","code","city","created_at"}).
            AddRow(1,"CGK","Jakarta", now).AddRow(2,"DPS","Denpasar", now))
    items, err := repo.List(context.Background(), 2, 0)
    if err != nil || len(items) != 2 { t.Fatalf("list err=%v n=%d", err, len(items)) }

    mock.ExpectExec(regexp.QuoteMeta(`UPDATE airports SET city=$2 WHERE code=$1`)).
        WithArgs("CGK", "NewCity").
        WillReturnResult(sqlmock.NewResult(0, 1))
    if err := repo.Update(context.Background(), "CGK", "NewCity"); err != nil { t.Fatalf("update: %v", err) }

    mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM airports WHERE code=$1`)).
        WithArgs("CGK").
        WillReturnResult(sqlmock.NewResult(0, 1))
    if err := repo.Delete(context.Background(), "CGK"); err != nil { t.Fatalf("delete: %v", err) }
}

func TestAirportRepo_GetByCode_Success_UpdateDeleteNotFound(t *testing.T) {
    db, mock, cleanup := newMockDB(t)
    defer cleanup()
    repo := NewAirportRepository(db)
    now := time.Now()
    mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, code, city, created_at FROM airports WHERE code=$1`)).
        WithArgs("CGK").
        WillReturnRows(sqlmock.NewRows([]string{"id","code","city","created_at"}).AddRow(1,"CGK","Jakarta", now))
    a, err := repo.GetByCode(context.Background(), "CGK")
    if err != nil || a.Code != "CGK" { t.Fatalf("get: %v a=%+v", err, a) }

    mock.ExpectExec(regexp.QuoteMeta(`UPDATE airports SET city=$2 WHERE code=$1`)).
        WithArgs("XXX", "City").
        WillReturnResult(sqlmock.NewResult(0, 0))
    if err := repo.Update(context.Background(), "XXX", "City"); err != domain.ErrAirportNotFound {
        t.Fatalf("want not found on update, got %v", err)
    }

    mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM airports WHERE code=$1`)).
        WithArgs("XXX").
        WillReturnResult(sqlmock.NewResult(0, 0))
    if err := repo.Delete(context.Background(), "XXX"); err != domain.ErrAirportNotFound {
        t.Fatalf("want not found on delete, got %v", err)
    }
}

// pqError is a minimal stub to simulate unique violation error messages.
type pqError struct{ msg string }
func (e *pqError) Error() string { return e.msg }
