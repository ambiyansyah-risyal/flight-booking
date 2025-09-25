package sqlxrepo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/ambiyansyah-risyal/flight-booking/internal/domain"
	"github.com/jmoiron/sqlx"
)

// ScheduleRepository stores flight schedules using sqlx.
type ScheduleRepository struct {
	db *sqlx.DB
}

func NewScheduleRepository(db *sqlx.DB) *ScheduleRepository {
	return &ScheduleRepository{db: db}
}

func (r *ScheduleRepository) Create(ctx context.Context, sched *domain.FlightSchedule) error {
	departure, err := time.Parse("2006-01-02", sched.DepartureDate)
	if err != nil {
		return domain.ErrInvalidScheduleDate
	}
	query := `INSERT INTO flight_schedules (route_code, airplane_code, departure_date) VALUES ($1, $2, $3) RETURNING id, departure_date, created_at`
	var storedDate, createdAt time.Time
	if err := r.db.QueryRowContext(ctx, query, sched.RouteCode, sched.AirplaneCode, departure).Scan(&sched.ID, &storedDate, &createdAt); err != nil {
		if isUniqueViolation(err) {
			return domain.ErrScheduleExists
		}
		return err
	}
	sched.DepartureDate = storedDate.Format("2006-01-02")
	sched.CreatedAt = createdAt.Format(time.RFC3339)
	return nil
}

func (r *ScheduleRepository) GetByID(ctx context.Context, id int64) (*domain.FlightSchedule, error) {
	row := r.db.QueryRowxContext(ctx, `SELECT id, route_code, airplane_code, departure_date, created_at FROM flight_schedules WHERE id=$1`, id)
	var sched domain.FlightSchedule
	var departure, createdAt time.Time
	if err := row.Scan(&sched.ID, &sched.RouteCode, &sched.AirplaneCode, &departure, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrScheduleNotFound
		}
		return nil, err
	}
	sched.DepartureDate = departure.Format("2006-01-02")
	sched.CreatedAt = createdAt.Format(time.RFC3339)
	return &sched, nil
}

func (r *ScheduleRepository) List(ctx context.Context, routeCode string, limit, offset int) ([]domain.FlightSchedule, error) {
	var (
		rows *sqlx.Rows
		err  error
	)
	if routeCode != "" {
		rows, err = r.db.QueryxContext(ctx, `SELECT id, route_code, airplane_code, departure_date, created_at FROM flight_schedules WHERE route_code=$1 ORDER BY departure_date LIMIT $2 OFFSET $3`, routeCode, limit, offset)
	} else {
		rows, err = r.db.QueryxContext(ctx, `SELECT id, route_code, airplane_code, departure_date, created_at FROM flight_schedules ORDER BY departure_date LIMIT $1 OFFSET $2`, limit, offset)
	}
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var items []domain.FlightSchedule
	for rows.Next() {
		var s domain.FlightSchedule
		var departure, createdAt time.Time
		if err := rows.Scan(&s.ID, &s.RouteCode, &s.AirplaneCode, &departure, &createdAt); err != nil {
			return nil, err
		}
		s.DepartureDate = departure.Format("2006-01-02")
		s.CreatedAt = createdAt.Format(time.RFC3339)
		items = append(items, s)
	}
	return items, rows.Err()
}

func (r *ScheduleRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM flight_schedules WHERE id=$1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrScheduleNotFound
	}
	return nil
}
