package domain

import "context"

// AirportRepository defines storage operations for airports.
type AirportRepository interface {
    Create(ctx context.Context, a *Airport) error
    GetByCode(ctx context.Context, code string) (*Airport, error)
    List(ctx context.Context, limit, offset int) ([]Airport, error)
    Update(ctx context.Context, code string, city string) error
    Delete(ctx context.Context, code string) error
}

