package usecase

import (
    "context"
    "time"

    "github.com/ambiyansyah-risyal/flight-booking/internal/domain"
)

type AirplaneUsecase struct {
    repo    domain.AirplaneRepository
    timeout time.Duration
}

func NewAirplaneUsecase(r domain.AirplaneRepository) *AirplaneUsecase {
    return &AirplaneUsecase{repo: r, timeout: 5 * time.Second}
}

func (u *AirplaneUsecase) Create(ctx context.Context, code string, seats int) (*domain.Airplane, error) {
    a := &domain.Airplane{Code: code, SeatCapacity: seats}
    a.Normalize()
    if err := a.Validate(); err != nil { return nil, err }
    ctx, cancel := context.WithTimeout(ctx, u.timeout)
    defer cancel()
    if err := u.repo.Create(ctx, a); err != nil { return nil, err }
    return a, nil
}

func (u *AirplaneUsecase) List(ctx context.Context, limit, offset int) ([]domain.Airplane, error) {
    if limit <= 0 || limit > 500 { limit = 50 }
    if offset < 0 { offset = 0 }
    ctx, cancel := context.WithTimeout(ctx, u.timeout)
    defer cancel()
    return u.repo.List(ctx, limit, offset)
}

func (u *AirplaneUsecase) UpdateSeats(ctx context.Context, code string, seats int) error {
    a := domain.Airplane{Code: code, SeatCapacity: seats}
    a.Normalize()
    if err := a.Validate(); err != nil { return err }
    ctx, cancel := context.WithTimeout(ctx, u.timeout)
    defer cancel()
    return u.repo.UpdateSeats(ctx, a.Code, a.SeatCapacity)
}

func (u *AirplaneUsecase) Delete(ctx context.Context, code string) error {
    a := domain.Airplane{Code: code, SeatCapacity: 1}
    a.Normalize()
    if err := a.Validate(); err != nil { return err }
    ctx, cancel := context.WithTimeout(ctx, u.timeout)
    defer cancel()
    return u.repo.Delete(ctx, a.Code)
}

