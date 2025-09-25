package usecase

import (
    "context"
    "time"

    "github.com/ambiyansyah-risyal/flight-booking/internal/domain"
)

type AirportUsecase struct {
    repo   domain.AirportRepository
    timeout time.Duration
}

func NewAirportUsecase(repo domain.AirportRepository) *AirportUsecase {
    return &AirportUsecase{repo: repo, timeout: 5 * time.Second}
}

func (u *AirportUsecase) Create(ctx context.Context, code, city string) (*domain.Airport, error) {
    a := &domain.Airport{Code: code, City: city}
    a.Normalize()
    if err := a.Validate(); err != nil {
        return nil, err
    }
    ctx, cancel := context.WithTimeout(ctx, u.timeout)
    defer cancel()
    if err := u.repo.Create(ctx, a); err != nil {
        return nil, err
    }
    return a, nil
}

func (u *AirportUsecase) List(ctx context.Context, limit, offset int) ([]domain.Airport, error) {
    if limit <= 0 || limit > 500 { limit = 50 }
    if offset < 0 { offset = 0 }
    ctx, cancel := context.WithTimeout(ctx, u.timeout)
    defer cancel()
    return u.repo.List(ctx, limit, offset)
}

func (u *AirportUsecase) Update(ctx context.Context, code, city string) error {
    a := domain.Airport{Code: code, City: city}
    a.Normalize()
    if err := a.Validate(); err != nil { return err }
    ctx, cancel := context.WithTimeout(ctx, u.timeout)
    defer cancel()
    return u.repo.Update(ctx, a.Code, a.City)
}

func (u *AirportUsecase) Delete(ctx context.Context, code string) error {
    a := domain.Airport{Code: code, City: "x"}
    a.Normalize()
    if err := a.Validate(); err != nil { return err }
    ctx, cancel := context.WithTimeout(ctx, u.timeout)
    defer cancel()
    return u.repo.Delete(ctx, a.Code)
}

