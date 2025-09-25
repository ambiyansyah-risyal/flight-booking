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

func newMockBookingDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	return sqlx.NewDb(db, "pgx"), mock, func() { _ = db.Close() }
}

func TestBookingRepository_Create_List_Get(t *testing.T) {
	db, mock, cleanup := newMockBookingDB(t)
	defer cleanup()
	repo := NewBookingRepository(db)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO bookings (reference, schedule_id, passenger_name, seat_number, status) VALUES ($1,$2,$3,$4,$5) RETURNING id, created_at`)).
		WithArgs("BK-AAAAAA", int64(1), "Alice", 1, domain.BookingStatusConfirmed).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(1, now))

	booking := &domain.Booking{Reference: "BK-AAAAAA", ScheduleID: 1, PassengerName: "Alice", SeatNumber: 1, Status: domain.BookingStatusConfirmed}
	if err := repo.Create(context.Background(), booking); err != nil {
		t.Fatalf("create: %v", err)
	}
	if booking.ID != 1 {
		t.Fatalf("expected id 1, got %d", booking.ID)
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM bookings WHERE schedule_id=$1`)).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	count, err := repo.CountBySchedule(context.Background(), 1)
	if err != nil || count != 1 {
		t.Fatalf("count: err=%v count=%d", err, count)
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, reference, schedule_id, passenger_name, seat_number, status, created_at FROM bookings WHERE schedule_id=$1 ORDER BY seat_number LIMIT $2 OFFSET $3`)).
		WithArgs(int64(1), 50, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "reference", "schedule_id", "passenger_name", "seat_number", "status", "created_at"}).
			AddRow(1, "BK-AAAAAA", 1, "Alice", 1, domain.BookingStatusConfirmed, now))
	list, err := repo.ListBySchedule(context.Background(), 1, 50, 0)
	if err != nil || len(list) != 1 {
		t.Fatalf("list: err=%v len=%d", err, len(list))
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, reference, schedule_id, passenger_name, seat_number, status, created_at FROM bookings WHERE reference=$1`)).
		WithArgs("BK-AAAAAA").
		WillReturnRows(sqlmock.NewRows([]string{"id", "reference", "schedule_id", "passenger_name", "seat_number", "status", "created_at"}).
			AddRow(1, "BK-AAAAAA", 1, "Alice", 1, domain.BookingStatusConfirmed, now))
	got, err := repo.GetByReference(context.Background(), "BK-AAAAAA")
	if err != nil || got.Reference != "BK-AAAAAA" {
		t.Fatalf("get: err=%v got=%+v", err, got)
	}
}

func TestBookingRepository_Errors(t *testing.T) {
	db, mock, cleanup := newMockBookingDB(t)
	defer cleanup()
	repo := NewBookingRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO bookings (reference, schedule_id, passenger_name, seat_number, status) VALUES ($1,$2,$3,$4,$5) RETURNING id, created_at`)).
		WithArgs("BK-AAAAAA", int64(1), "Alice", 1, domain.BookingStatusConfirmed).
		WillReturnError(&pqErr{msg: "duplicate key value violates unique constraint"})
	if err := repo.Create(context.Background(), &domain.Booking{Reference: "BK-AAAAAA", ScheduleID: 1, PassengerName: "Alice", SeatNumber: 1, Status: domain.BookingStatusConfirmed}); err != domain.ErrBookingExists {
		t.Fatalf("want exists, got %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM bookings WHERE schedule_id=$1`)).
		WithArgs(int64(2)).
		WillReturnError(fmt.Errorf("db error"))
	if _, err := repo.CountBySchedule(context.Background(), 2); err == nil {
		t.Fatalf("expected count error")
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, reference, schedule_id, passenger_name, seat_number, status, created_at FROM bookings WHERE reference=$1`)).
		WithArgs("BK-NOTFOUND").
		WillReturnRows(sqlmock.NewRows([]string{"id", "reference", "schedule_id", "passenger_name", "seat_number", "status", "created_at"}))
	if _, err := repo.GetByReference(context.Background(), "BK-NOTFOUND"); err != domain.ErrBookingNotFound {
		t.Fatalf("want not found, got %v", err)
	}
}
