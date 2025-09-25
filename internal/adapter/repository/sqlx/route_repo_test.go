package sqlxrepo

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ambiyansyah-risyal/flight-booking/internal/domain"
)

func TestRouteRepository_Create_Get_List_Delete(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewRouteRepository(db)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO routes (code, origin_code, destination_code) VALUES ($1, $2, $3) RETURNING id, created_at`)).
		WithArgs("RT1", "CGK", "DPS").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(1, now))
	r := &domain.Route{Code: "RT1", OriginCode: "CGK", DestinationCode: "DPS"}
	if err := repo.Create(context.Background(), r); err != nil {
		t.Fatalf("create: %v", err)
	}
	if r.ID != 1 || r.CreatedAt == "" {
		t.Fatalf("route fields not set: %+v", r)
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, code, origin_code, destination_code, created_at FROM routes WHERE code=$1`)).
		WithArgs("RT1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "origin_code", "destination_code", "created_at"}).AddRow(1, "RT1", "CGK", "DPS", now))
	got, err := repo.GetByCode(context.Background(), "RT1")
	if err != nil || got.Code != "RT1" {
		t.Fatalf("get: err=%v route=%+v", err, got)
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, code, origin_code, destination_code, created_at FROM routes ORDER BY code LIMIT $1 OFFSET $2`)).
		WithArgs(10, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "origin_code", "destination_code", "created_at"}).AddRow(1, "RT1", "CGK", "DPS", now))
	list, err := repo.List(context.Background(), 10, 0)
	if err != nil || len(list) != 1 {
		t.Fatalf("list: err=%v len=%d", err, len(list))
	}

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM routes WHERE code=$1`)).
		WithArgs("RT1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.Delete(context.Background(), "RT1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestRouteRepository_CreateDuplicate_GetNotFound_DeleteNotFound(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewRouteRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO routes (code, origin_code, destination_code) VALUES ($1, $2, $3) RETURNING id, created_at`)).
		WithArgs("RT1", "CGK", "DPS").
		WillReturnError(&pqError{msg: "duplicate key value violates unique constraint"})
	if err := repo.Create(context.Background(), &domain.Route{Code: "RT1", OriginCode: "CGK", DestinationCode: "DPS"}); err != domain.ErrRouteExists {
		t.Fatalf("want ErrRouteExists, got %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, code, origin_code, destination_code, created_at FROM routes WHERE code=$1`)).
		WithArgs("NONE").
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "origin_code", "destination_code", "created_at"}))
	if _, err := repo.GetByCode(context.Background(), "NONE"); err != domain.ErrRouteNotFound {
		t.Fatalf("want ErrRouteNotFound, got %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM routes WHERE code=$1`)).
		WithArgs("NONE").
		WillReturnResult(sqlmock.NewResult(0, 0))
	if err := repo.Delete(context.Background(), "NONE"); err != domain.ErrRouteNotFound {
		t.Fatalf("want ErrRouteNotFound on delete, got %v", err)
	}
}

func TestRouteRepository_ListErrors(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewRouteRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, code, origin_code, destination_code, created_at FROM routes ORDER BY code LIMIT $1 OFFSET $2`)).
		WithArgs(5, 0).
		WillReturnError(errors.New("db down"))
	if _, err := repo.List(context.Background(), 5, 0); err == nil {
		t.Fatalf("expected list error")
	}
}
