package usecase

import (
	"context"
	"time"

	"github.com/ambiyansyah-risyal/flight-booking/internal/domain"
)

// RouteUsecase orchestrates business rules around managing flight routes.
type RouteUsecase struct {
	routes   domain.RouteRepository
	airports domain.AirportRepository
	timeout  time.Duration
}

// NewRouteUsecase creates a new route usecase with sensible defaults.
func NewRouteUsecase(routeRepo domain.RouteRepository, airportRepo domain.AirportRepository) *RouteUsecase {
	return &RouteUsecase{routes: routeRepo, airports: airportRepo, timeout: 5 * time.Second}
}

// Create stores a new route after validating the payload and ensuring airports exist.
func (u *RouteUsecase) Create(ctx context.Context, code, origin, destination string) (*domain.Route, error) {
	r := &domain.Route{Code: code, OriginCode: origin, DestinationCode: destination}
	r.Normalize()
	if err := r.Validate(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()

	if _, err := u.airports.GetByCode(ctx, r.OriginCode); err != nil {
		return nil, err
	}
	if _, err := u.airports.GetByCode(ctx, r.DestinationCode); err != nil {
		return nil, err
	}
	if err := u.routes.Create(ctx, r); err != nil {
		return nil, err
	}
	return r, nil
}

// List returns paginated routes sorted by code.
func (u *RouteUsecase) List(ctx context.Context, limit, offset int) ([]domain.Route, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	ctx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()
	return u.routes.List(ctx, limit, offset)
}

// Delete removes a route by code.
func (u *RouteUsecase) Delete(ctx context.Context, code string) error {
	r := domain.Route{Code: code, OriginCode: "X", DestinationCode: "Y"}
	r.Normalize()
	if err := r.Validate(); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()
	return u.routes.Delete(ctx, r.Code)
}
