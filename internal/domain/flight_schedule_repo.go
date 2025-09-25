package domain

import "context"

// FlightScheduleRepository handles persistence for scheduled flights.
type FlightScheduleRepository interface {
	Create(ctx context.Context, s *FlightSchedule) error
	GetByID(ctx context.Context, id int64) (*FlightSchedule, error)
	List(ctx context.Context, routeCode string, limit, offset int) ([]FlightSchedule, error)
	Delete(ctx context.Context, id int64) error
}
