package domain

import "context"

type AirplaneRepository interface {
    Create(ctx context.Context, a *Airplane) error
    GetByCode(ctx context.Context, code string) (*Airplane, error)
    List(ctx context.Context, limit, offset int) ([]Airplane, error)
    UpdateSeats(ctx context.Context, code string, seats int) error
    Delete(ctx context.Context, code string) error
}

