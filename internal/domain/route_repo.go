package domain

import "context"

// RouteRepository defines storage operations for flight routes.
type RouteRepository interface {
	Create(ctx context.Context, r *Route) error
	GetByCode(ctx context.Context, code string) (*Route, error)
	List(ctx context.Context, limit, offset int) ([]Route, error)
	Delete(ctx context.Context, code string) error
}
