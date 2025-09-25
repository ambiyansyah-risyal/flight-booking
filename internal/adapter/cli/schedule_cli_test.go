package cli

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ambiyansyah-risyal/flight-booking/internal/domain"
	"github.com/jmoiron/sqlx"
)

type fakeScheduleRepoCLI struct {
	nextID int64
	items  map[int64]domain.FlightSchedule
}

func (f *fakeScheduleRepoCLI) GetByID(ctx context.Context, id int64) (*domain.FlightSchedule, error) {
	if f.items == nil {
		f.items = make(map[int64]domain.FlightSchedule)
	}
	if s, ok := f.items[id]; ok {
		copy := s
		return &copy, nil
	}
	return nil, domain.ErrScheduleNotFound
}

func (f *fakeScheduleRepoCLI) Create(ctx context.Context, s *domain.FlightSchedule) error {
	if f.items == nil {
		f.items = make(map[int64]domain.FlightSchedule)
	}
	f.nextID++
	s.ID = f.nextID
	f.items[s.ID] = *s
	return nil
}

func (f *fakeScheduleRepoCLI) List(ctx context.Context, routeCode string, limit, offset int) ([]domain.FlightSchedule, error) {
	var out []domain.FlightSchedule
	for _, item := range f.items {
		if routeCode == "" || item.RouteCode == routeCode {
			out = append(out, item)
		}
	}
	return out, nil
}

func (f *fakeScheduleRepoCLI) Delete(ctx context.Context, id int64) error {
	if f.items == nil {
		f.items = make(map[int64]domain.FlightSchedule)
	}
	if _, ok := f.items[id]; !ok {
		return domain.ErrScheduleNotFound
	}
	delete(f.items, id)
	return nil
}

type fakeRouteRepoCLIForSchedule struct{ existing map[string]bool }

func (f *fakeRouteRepoCLIForSchedule) Create(ctx context.Context, r *domain.Route) error { return nil }
func (f *fakeRouteRepoCLIForSchedule) GetByCode(ctx context.Context, code string) (*domain.Route, error) {
	if f.existing != nil && f.existing[code] {
		return &domain.Route{Code: code}, nil
	}
	return nil, domain.ErrRouteNotFound
}
func (f *fakeRouteRepoCLIForSchedule) List(ctx context.Context, limit, offset int) ([]domain.Route, error) {
	return nil, nil
}
func (f *fakeRouteRepoCLIForSchedule) Delete(ctx context.Context, code string) error { return nil }

type fakeAirplaneRepoCLIForSchedule struct{ existing map[string]bool }

func (f *fakeAirplaneRepoCLIForSchedule) Create(ctx context.Context, a *domain.Airplane) error {
	return nil
}
func (f *fakeAirplaneRepoCLIForSchedule) GetByCode(ctx context.Context, code string) (*domain.Airplane, error) {
	if f.existing != nil && f.existing[code] {
		return &domain.Airplane{Code: code}, nil
	}
	return nil, domain.ErrAirplaneNotFound
}
func (f *fakeAirplaneRepoCLIForSchedule) List(ctx context.Context, limit, offset int) ([]domain.Airplane, error) {
	return nil, nil
}
func (f *fakeAirplaneRepoCLIForSchedule) UpdateSeats(ctx context.Context, code string, seats int) error {
	return nil
}
func (f *fakeAirplaneRepoCLIForSchedule) Delete(ctx context.Context, code string) error { return nil }

func TestScheduleCLI_Flow(t *testing.T) {
	oldDB, oldRepo, oldRouteRepo, oldPlaneRepo := newScheduleDB, newScheduleRepo, newScheduleRouteRepo, newScheduleAirplaneRepo
	t.Cleanup(func() {
		newScheduleDB = oldDB
		newScheduleRepo = oldRepo
		newScheduleRouteRepo = oldRouteRepo
		newScheduleAirplaneRepo = oldPlaneRepo
	})
	newScheduleDB = func(string) (*sqlx.DB, error) {
		db, _, err := sqlmock.New()
		if err != nil {
			return nil, fmt.Errorf("sqlmock: %w", err)
		}
		return sqlx.NewDb(db, "pgx"), nil
	}
	schedules := &fakeScheduleRepoCLI{}
	routes := &fakeRouteRepoCLIForSchedule{existing: map[string]bool{"RT1": true}}
	planes := &fakeAirplaneRepoCLIForSchedule{existing: map[string]bool{"A320": true}}
	newScheduleRepo = func(*sqlx.DB) domain.FlightScheduleRepository { return schedules }
	newScheduleRouteRepo = func(*sqlx.DB) domain.RouteRepository { return routes }
	newScheduleAirplaneRepo = func(*sqlx.DB) domain.AirplaneRepository { return planes }

	t.Setenv("FLIGHT_DB_HOST", "localhost")

	os.Args = []string{"flight-booking", "schedule", "create", "--route", "RT1", "--airplane", "A320", "--date", "2025-01-02"}
	if err := Execute(); err != nil {
		t.Fatalf("create: %v", err)
	}

	os.Args = []string{"flight-booking", "schedule", "list", "--route", "RT1"}
	if err := Execute(); err != nil {
		t.Fatalf("list: %v", err)
	}

	os.Args = []string{"flight-booking", "schedule", "delete", "1"}
	if err := Execute(); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestScheduleCLI_CreateMissingFlags(t *testing.T) {
	t.Setenv("FLIGHT_DB_HOST", "localhost")
	os.Args = []string{"flight-booking", "schedule", "create"}
	if err := Execute(); err == nil {
		t.Fatalf("expected flag error")
	}
}

func TestScheduleCLI_DeleteInvalidID(t *testing.T) {
	t.Setenv("FLIGHT_DB_HOST", "localhost")
	os.Args = []string{"flight-booking", "schedule", "delete", "bad"}
	if err := Execute(); err == nil {
		t.Fatalf("expected parse error")
	}
}

func TestScheduleCLI_DeleteNotFound(t *testing.T) {
	oldDB, oldRepo, oldRouteRepo, oldPlaneRepo := newScheduleDB, newScheduleRepo, newScheduleRouteRepo, newScheduleAirplaneRepo
	t.Cleanup(func() {
		newScheduleDB = oldDB
		newScheduleRepo = oldRepo
		newScheduleRouteRepo = oldRouteRepo
		newScheduleAirplaneRepo = oldPlaneRepo
	})
	newScheduleDB = func(string) (*sqlx.DB, error) {
		db, _, _ := sqlmock.New()
		return sqlx.NewDb(db, "pgx"), nil
	}
	schedules := &fakeScheduleRepoCLI{}
	routes := &fakeRouteRepoCLIForSchedule{existing: map[string]bool{"RT1": true}}
	planes := &fakeAirplaneRepoCLIForSchedule{existing: map[string]bool{"A320": true}}
	newScheduleRepo = func(*sqlx.DB) domain.FlightScheduleRepository { return schedules }
	newScheduleRouteRepo = func(*sqlx.DB) domain.RouteRepository { return routes }
	newScheduleAirplaneRepo = func(*sqlx.DB) domain.AirplaneRepository { return planes }

	t.Setenv("FLIGHT_DB_HOST", "localhost")
	os.Args = []string{"flight-booking", "schedule", "delete", "1"}
	if err := Execute(); err == nil {
		t.Fatalf("expected delete error")
	}
}
