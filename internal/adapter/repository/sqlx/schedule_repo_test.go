package sqlxrepo

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ambiyansyah-risyal/flight-booking/internal/domain"
)

func TestScheduleRepository_Create_List_Delete(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewScheduleRepository(db)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO flight_schedules (route_code, airplane_code, departure_date) VALUES ($1, $2, $3) RETURNING id, departure_date, created_at`)).
		WithArgs("RT1", "A320", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "departure_date", "created_at"}).AddRow(1, now, now))
	sched := &domain.FlightSchedule{RouteCode: "RT1", AirplaneCode: "A320", DepartureDate: "2025-01-02"}
	if err := repo.Create(context.Background(), sched); err != nil {
		t.Fatalf("create: %v", err)
	}
	if sched.ID != 1 || sched.CreatedAt == "" {
		t.Fatalf("schedule fields not set: %+v", sched)
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, route_code, airplane_code, departure_date, created_at FROM flight_schedules WHERE route_code=$1 ORDER BY departure_date LIMIT $2 OFFSET $3`)).
		WithArgs("RT1", 10, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "route_code", "airplane_code", "departure_date", "created_at"}).AddRow(1, "RT1", "A320", now, now))
	list, err := repo.List(context.Background(), "RT1", 10, 0)
	if err != nil || len(list) != 1 {
		t.Fatalf("list err=%v len=%d", err, len(list))
	}

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM flight_schedules WHERE id=$1`)).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.Delete(context.Background(), 1); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestScheduleRepository_CreateDuplicate_InvalidDate_DeleteNotFound(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewScheduleRepository(db)

	if err := repo.Create(context.Background(), &domain.FlightSchedule{RouteCode: "RT1", AirplaneCode: "A320", DepartureDate: "bad"}); err != domain.ErrInvalidScheduleDate {
		t.Fatalf("want invalid date, got %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO flight_schedules (route_code, airplane_code, departure_date) VALUES ($1, $2, $3) RETURNING id, departure_date, created_at`)).
		WithArgs("RT1", "A320", sqlmock.AnyArg()).
		WillReturnError(&pqError{msg: "duplicate key value violates unique constraint"})
	if err := repo.Create(context.Background(), &domain.FlightSchedule{RouteCode: "RT1", AirplaneCode: "A320", DepartureDate: "2025-01-02"}); err != domain.ErrScheduleExists {
		t.Fatalf("want schedule exists, got %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM flight_schedules WHERE id=$1`)).
		WithArgs(int64(99)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	if err := repo.Delete(context.Background(), 99); err != domain.ErrScheduleNotFound {
		t.Fatalf("want schedule not found, got %v", err)
	}
}

func TestScheduleRepository_GetByID(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewScheduleRepository(db)
	now := time.Now()

	// Test successful retrieval
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, route_code, airplane_code, departure_date, created_at FROM flight_schedules WHERE id=$1`)).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "route_code", "airplane_code", "departure_date", "created_at"}).
			AddRow(1, "R1", "A1", now, now))
	
	sched, err := repo.GetByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if sched.ID != 1 || sched.RouteCode != "R1" {
		t.Fatalf("schedule fields not set correctly: %+v", sched)
	}

	// Test not found
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, route_code, airplane_code, departure_date, created_at FROM flight_schedules WHERE id=$1`)).
		WithArgs(int64(99)).
		WillReturnError(sql.ErrNoRows)
	sched, err = repo.GetByID(context.Background(), 99)
	if err != domain.ErrScheduleNotFound {
		t.Fatalf("want schedule not found, got %v", err)
	}
	if sched != nil {
		t.Fatalf("expected nil schedule for not found")
	}
}

func TestScheduleRepository_ListErrors(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewScheduleRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, route_code, airplane_code, departure_date, created_at FROM flight_schedules ORDER BY departure_date LIMIT $1 OFFSET $2`)).
		WithArgs(5, 0).
		WillReturnError(errors.New("db down"))
	if _, err := repo.List(context.Background(), "", 5, 0); err == nil {
		t.Fatalf("expected list error")
	}
}
