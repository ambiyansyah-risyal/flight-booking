package usecase

import (
	"context"
	"testing"

	"github.com/ambiyansyah-risyal/flight-booking/internal/domain"
)

type fakeRouteRepo struct {
	items     map[string]domain.Route
	createErr error
	deleteErr error
}

func (f *fakeRouteRepo) Create(ctx context.Context, r *domain.Route) error {
	if f.createErr != nil {
		return f.createErr
	}
	if f.items == nil {
		f.items = make(map[string]domain.Route)
	}
	f.items[r.Code] = *r
	return nil
}

func (f *fakeRouteRepo) GetByCode(ctx context.Context, code string) (*domain.Route, error) {
	if r, ok := f.items[code]; ok {
		return &r, nil
	}
	return nil, domain.ErrRouteNotFound
}

func (f *fakeRouteRepo) List(ctx context.Context, limit, offset int) ([]domain.Route, error) {
	var out []domain.Route
	for _, r := range f.items {
		out = append(out, r)
	}
	return out, nil
}

func (f *fakeRouteRepo) Delete(ctx context.Context, code string) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	if f.items == nil {
		f.items = make(map[string]domain.Route)
	}
	if _, ok := f.items[code]; !ok {
		return domain.ErrRouteNotFound
	}
	delete(f.items, code)
	return nil
}

type fakeAirportRepoRoute struct{ existing map[string]bool }

func (f *fakeAirportRepoRoute) Create(ctx context.Context, a *domain.Airport) error { return nil }
func (f *fakeAirportRepoRoute) GetByCode(ctx context.Context, code string) (*domain.Airport, error) {
	if f.existing != nil && f.existing[code] {
		return &domain.Airport{Code: code}, nil
	}
	return nil, domain.ErrAirportNotFound
}
func (f *fakeAirportRepoRoute) List(ctx context.Context, limit, offset int) ([]domain.Airport, error) {
	return nil, nil
}
func (f *fakeAirportRepoRoute) Update(ctx context.Context, code string, city string) error {
	return nil
}
func (f *fakeAirportRepoRoute) Delete(ctx context.Context, code string) error { return nil }

func TestRouteUsecase_Create_List_Delete(t *testing.T) {
	rr := &fakeRouteRepo{items: make(map[string]domain.Route)}
	ar := &fakeAirportRepoRoute{existing: map[string]bool{"CGK": true, "DPS": true}}
	uc := NewRouteUsecase(rr, ar)
	if _, err := uc.Create(context.Background(), " rt1 ", " cgk ", " dps "); err != nil {
		t.Fatalf("create: %v", err)
	}
	items, err := uc.List(context.Background(), -1, -1)
	if err != nil || len(items) != 1 {
		t.Fatalf("list: err=%v len=%d", err, len(items))
	}
	if err := uc.Delete(context.Background(), "rt1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestRouteUsecase_Create_AirportMissing(t *testing.T) {
	rr := &fakeRouteRepo{}
	ar := &fakeAirportRepoRoute{existing: map[string]bool{"CGK": true}}
	uc := NewRouteUsecase(rr, ar)
	if _, err := uc.Create(context.Background(), "R1", "CGK", "DPS"); err != domain.ErrAirportNotFound {
		t.Fatalf("want airport not found, got %v", err)
	}
}

func TestRouteUsecase_Delete_InvalidCode(t *testing.T) {
	rr := &fakeRouteRepo{}
	ar := &fakeAirportRepoRoute{}
	uc := NewRouteUsecase(rr, ar)
	if err := uc.Delete(context.Background(), ""); err != domain.ErrInvalidRouteCode {
		t.Fatalf("want invalid route code, got %v", err)
	}
}
