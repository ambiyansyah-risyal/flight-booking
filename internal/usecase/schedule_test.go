package usecase

import (
	"context"
	"testing"

	"github.com/ambiyansyah-risyal/flight-booking/internal/domain"
)

type fakeScheduleRepo struct {
	items     []domain.FlightSchedule
	createErr error
	deleteErr error
}

func (f *fakeScheduleRepo) Create(ctx context.Context, s *domain.FlightSchedule) error {
	if f.createErr != nil {
		return f.createErr
	}
	s.ID = int64(len(f.items) + 1)
	f.items = append(f.items, *s)
	return nil
}

func (f *fakeScheduleRepo) GetByID(ctx context.Context, id int64) (*domain.FlightSchedule, error) {
	for i := range f.items {
		if f.items[i].ID == id {
			item := f.items[i]
			return &item, nil
		}
	}
	return nil, domain.ErrScheduleNotFound
}

func (f *fakeScheduleRepo) List(ctx context.Context, routeCode string, limit, offset int) ([]domain.FlightSchedule, error) {
	var out []domain.FlightSchedule
	for _, item := range f.items {
		if routeCode == "" || item.RouteCode == routeCode {
			out = append(out, item)
		}
	}
	return out, nil
}

func (f *fakeScheduleRepo) Delete(ctx context.Context, id int64) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	for i, item := range f.items {
		if item.ID == id {
			f.items = append(f.items[:i], f.items[i+1:]...)
			return nil
		}
	}
	return domain.ErrScheduleNotFound
}

type fakeRouteRepoSched struct{ items map[string]bool }

func (f *fakeRouteRepoSched) Create(ctx context.Context, r *domain.Route) error { return nil }
func (f *fakeRouteRepoSched) GetByCode(ctx context.Context, code string) (*domain.Route, error) {
	if f.items != nil && f.items[code] {
		return &domain.Route{Code: code}, nil
	}
	return nil, domain.ErrRouteNotFound
}
func (f *fakeRouteRepoSched) List(ctx context.Context, limit, offset int) ([]domain.Route, error) {
	return nil, nil
}
func (f *fakeRouteRepoSched) Delete(ctx context.Context, code string) error { return nil }

type fakeAirplaneRepoSched struct{ items map[string]bool }

func (f *fakeAirplaneRepoSched) Create(ctx context.Context, a *domain.Airplane) error { return nil }
func (f *fakeAirplaneRepoSched) GetByCode(ctx context.Context, code string) (*domain.Airplane, error) {
	if f.items != nil && f.items[code] {
		return &domain.Airplane{Code: code}, nil
	}
	return nil, domain.ErrAirplaneNotFound
}
func (f *fakeAirplaneRepoSched) List(ctx context.Context, limit, offset int) ([]domain.Airplane, error) {
	return nil, nil
}
func (f *fakeAirplaneRepoSched) UpdateSeats(ctx context.Context, code string, seats int) error {
	return nil
}
func (f *fakeAirplaneRepoSched) Delete(ctx context.Context, code string) error { return nil }

func TestScheduleUsecase_Create_List_Delete(t *testing.T) {
	repo := &fakeScheduleRepo{}
	routes := &fakeRouteRepoSched{items: map[string]bool{"RT1": true}}
	planes := &fakeAirplaneRepoSched{items: map[string]bool{"A320": true}}
	uc := NewScheduleUsecase(repo, routes, planes)
	sched, err := uc.Create(context.Background(), " rt1 ", " a320 ", "2025-01-02")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if sched.ID == 0 {
		t.Fatalf("expected id assigned, got %d", sched.ID)
	}
	items, err := uc.List(context.Background(), "rt1", 10, 0)
	if err != nil || len(items) != 1 {
		t.Fatalf("list err=%v len=%d", err, len(items))
	}
	if err := uc.Delete(context.Background(), sched.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestScheduleUsecase_Create_MissingReferences(t *testing.T) {
	repo := &fakeScheduleRepo{}
	routes := &fakeRouteRepoSched{items: map[string]bool{"RT1": true}}
	planes := &fakeAirplaneRepoSched{}
	uc := NewScheduleUsecase(repo, routes, planes)
	if _, err := uc.Create(context.Background(), "RT1", "A320", "2025-01-01"); err != domain.ErrAirplaneNotFound {
		t.Fatalf("want airplane not found, got %v", err)
	}
	planes.items = map[string]bool{"A320": true}
	routes.items = map[string]bool{}
	if _, err := uc.Create(context.Background(), "RT1", "A320", "2025-01-01"); err != domain.ErrRouteNotFound {
		t.Fatalf("want route not found, got %v", err)
	}
}

func TestScheduleUsecase_Delete_InvalidID(t *testing.T) {
	repo := &fakeScheduleRepo{}
	uc := NewScheduleUsecase(repo, &fakeRouteRepoSched{}, &fakeAirplaneRepoSched{})
	if err := uc.Delete(context.Background(), 0); err != domain.ErrInvalidScheduleID {
		t.Fatalf("want invalid schedule id, got %v", err)
	}
}
