package sqlxrepo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/ambiyansyah-risyal/flight-booking/internal/domain"
	"github.com/jmoiron/sqlx"
)

// BookingRepository persists bookings via sqlx.
type BookingRepository struct {
	db *sqlx.DB
}

func NewBookingRepository(db *sqlx.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

func (r *BookingRepository) Create(ctx context.Context, b *domain.Booking) error {
	query := `INSERT INTO bookings (reference, schedule_id, passenger_name, seat_number, status) VALUES ($1,$2,$3,$4,$5) RETURNING id, created_at`
	var createdAt time.Time
	if err := r.db.QueryRowContext(ctx, query, b.Reference, b.ScheduleID, b.PassengerName, b.SeatNumber, b.Status).Scan(&b.ID, &createdAt); err != nil {
		if isUniqueViolation(err) {
			return domain.ErrBookingExists
		}
		if isForeignKeyViolation(err) {
			return domain.ErrScheduleNotFound
		}
		return err
	}
	b.CreatedAt = createdAt.Format(time.RFC3339)
	return nil
}

func (r *BookingRepository) CountBySchedule(ctx context.Context, scheduleID int64) (int, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM bookings WHERE schedule_id=$1`, scheduleID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *BookingRepository) ListBySchedule(ctx context.Context, scheduleID int64, limit, offset int) ([]domain.Booking, error) {
	rows, err := r.db.QueryxContext(ctx, `SELECT id, reference, schedule_id, passenger_name, seat_number, status, created_at FROM bookings WHERE schedule_id=$1 ORDER BY seat_number LIMIT $2 OFFSET $3`, scheduleID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var items []domain.Booking
	for rows.Next() {
		var b domain.Booking
		var createdAt time.Time
		if err := rows.Scan(&b.ID, &b.Reference, &b.ScheduleID, &b.PassengerName, &b.SeatNumber, &b.Status, &createdAt); err != nil {
			return nil, err
		}
		b.CreatedAt = createdAt.Format(time.RFC3339)
		items = append(items, b)
	}
	return items, rows.Err()
}

func (r *BookingRepository) GetByReference(ctx context.Context, reference string) (*domain.Booking, error) {
	row := r.db.QueryRowxContext(ctx, `SELECT id, reference, schedule_id, passenger_name, seat_number, status, created_at FROM bookings WHERE reference=$1`, reference)
	var b domain.Booking
	var createdAt time.Time
	if err := row.Scan(&b.ID, &b.Reference, &b.ScheduleID, &b.PassengerName, &b.SeatNumber, &b.Status, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrBookingNotFound
		}
		return nil, err
	}
	b.CreatedAt = createdAt.Format(time.RFC3339)
	return &b, nil
}
