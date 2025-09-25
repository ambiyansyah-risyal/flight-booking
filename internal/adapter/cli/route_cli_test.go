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

type fakeRouteRepoCLI struct{ data map[string]domain.Route }

func (f *fakeRouteRepoCLI) Create(ctx context.Context, r *domain.Route) error {
	if f.data == nil {
		f.data = make(map[string]domain.Route)
	}
	if _, ok := f.data[r.Code]; ok {
		return domain.ErrRouteExists
	}
	f.data[r.Code] = *r
	return nil
}

func (f *fakeRouteRepoCLI) GetByCode(ctx context.Context, code string) (*domain.Route, error) {
	if r, ok := f.data[code]; ok {
		return &r, nil
	}
	return nil, domain.ErrRouteNotFound
}

func (f *fakeRouteRepoCLI) List(ctx context.Context, limit, offset int) ([]domain.Route, error) {
	var routes []domain.Route
	for _, r := range f.data {
		routes = append(routes, r)
	}
	return routes, nil
}

func (f *fakeRouteRepoCLI) Delete(ctx context.Context, code string) error {
	if f.data == nil {
		f.data = make(map[string]domain.Route)
	}
	if _, ok := f.data[code]; !ok {
		return domain.ErrRouteNotFound
	}
	delete(f.data, code)
	return nil
}

type fakeAirportRepoCLI struct{ existing map[string]bool }

func (f *fakeAirportRepoCLI) Create(ctx context.Context, a *domain.Airport) error { return nil }
func (f *fakeAirportRepoCLI) GetByCode(ctx context.Context, code string) (*domain.Airport, error) {
	if f.existing != nil && f.existing[code] {
		return &domain.Airport{Code: code}, nil
	}
	return nil, domain.ErrAirportNotFound
}
func (f *fakeAirportRepoCLI) List(ctx context.Context, limit, offset int) ([]domain.Airport, error) {
	return nil, nil
}
func (f *fakeAirportRepoCLI) Update(ctx context.Context, code string, city string) error { return nil }
func (f *fakeAirportRepoCLI) Delete(ctx context.Context, code string) error              { return nil }

func TestRouteCLI_Flow(t *testing.T) {
	oldDB, oldRouteRepo, oldAirportRepo := newRouteDB, newRouteRepo, newRouteAirportRepo
	t.Cleanup(func() {
		newRouteDB = oldDB
		newRouteRepo = oldRouteRepo
		newRouteAirportRepo = oldAirportRepo
	})
	newRouteDB = func(string) (*sqlx.DB, error) {
		db, _, err := sqlmock.New()
		if err != nil {
			return nil, fmt.Errorf("sqlmock: %w", err)
		}
		return sqlx.NewDb(db, "pgx"), nil
	}
	routes := &fakeRouteRepoCLI{data: map[string]domain.Route{}}
	airports := &fakeAirportRepoCLI{existing: map[string]bool{"CGK": true, "DPS": true}}
	newRouteRepo = func(*sqlx.DB) domain.RouteRepository { return routes }
	newRouteAirportRepo = func(*sqlx.DB) domain.AirportRepository { return airports }

	t.Setenv("FLIGHT_DB_HOST", "localhost")

	os.Args = []string{"flight-booking", "route", "create", "--code", "RT1", "--origin", "CGK", "--destination", "DPS"}
	if err := Execute(); err != nil {
		t.Fatalf("create: %v", err)
	}

	os.Args = []string{"flight-booking", "route", "list"}
	if err := Execute(); err != nil {
		t.Fatalf("list: %v", err)
	}

	os.Args = []string{"flight-booking", "route", "delete", "RT1"}
	if err := Execute(); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestRouteCLI_CreateMissingFlags(t *testing.T) {
	t.Setenv("FLIGHT_DB_HOST", "localhost")
	os.Args = []string{"flight-booking", "route", "create"}
	if err := Execute(); err == nil {
		t.Fatalf("expected missing flag error")
	}
}

func TestRouteCLI_DeleteNotFound(t *testing.T) {
	oldDB, oldRouteRepo, oldAirportRepo := newRouteDB, newRouteRepo, newRouteAirportRepo
	t.Cleanup(func() {
		newRouteDB = oldDB
		newRouteRepo = oldRouteRepo
		newRouteAirportRepo = oldAirportRepo
	})
	newRouteDB = func(string) (*sqlx.DB, error) {
		db, _, _ := sqlmock.New()
		return sqlx.NewDb(db, "pgx"), nil
	}
	routes := &fakeRouteRepoCLI{data: map[string]domain.Route{}}
	airports := &fakeAirportRepoCLI{existing: map[string]bool{"CGK": true, "DPS": true}}
	newRouteRepo = func(*sqlx.DB) domain.RouteRepository { return routes }
	newRouteAirportRepo = func(*sqlx.DB) domain.AirportRepository { return airports }

	t.Setenv("FLIGHT_DB_HOST", "localhost")
	os.Args = []string{"flight-booking", "route", "delete", "RT1"}
	if err := Execute(); err == nil {
		t.Fatalf("expected delete error")
	}
}
