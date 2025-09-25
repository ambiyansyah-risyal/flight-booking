package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/ambiyansyah-risyal/flight-booking/internal/domain"
)

// ScheduleUsecase coordinates validation for flight schedules.
type ScheduleUsecase struct {
	schedules domain.FlightScheduleRepository
	routes    domain.RouteRepository
	airplanes domain.AirplaneRepository
	timeout   time.Duration
}

// NewScheduleUsecase constructs a ScheduleUsecase with default timeout.
func NewScheduleUsecase(repo domain.FlightScheduleRepository, routeRepo domain.RouteRepository, airplaneRepo domain.AirplaneRepository) *ScheduleUsecase {
	return &ScheduleUsecase{schedules: repo, routes: routeRepo, airplanes: airplaneRepo, timeout: 5 * time.Second}
}

// Create validates references and stores a new flight schedule.
func (u *ScheduleUsecase) Create(ctx context.Context, routeCode, airplaneCode, departureDate string) (*domain.FlightSchedule, error) {
	sched := &domain.FlightSchedule{RouteCode: routeCode, AirplaneCode: airplaneCode, DepartureDate: departureDate}
	sched.Normalize()
	if err := sched.Validate(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()

	if _, err := u.routes.GetByCode(ctx, sched.RouteCode); err != nil {
		return nil, err
	}
	if _, err := u.airplanes.GetByCode(ctx, sched.AirplaneCode); err != nil {
		return nil, err
	}
	if err := u.schedules.Create(ctx, sched); err != nil {
		return nil, err
	}
	return sched, nil
}

// List returns upcoming schedules, optionally filtered by route code.
func (u *ScheduleUsecase) List(ctx context.Context, routeCode string, limit, offset int) ([]domain.FlightSchedule, error) {
	routeCode = strings.ToUpper(strings.TrimSpace(routeCode))
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	ctx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()
	return u.schedules.List(ctx, routeCode, limit, offset)
}

// Delete removes a schedule by identifier.
func (u *ScheduleUsecase) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return domain.ErrInvalidScheduleID
	}
	ctx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()
	return u.schedules.Delete(ctx, id)
}
