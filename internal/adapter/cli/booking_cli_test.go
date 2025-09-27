package cli

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ambiyansyah-risyal/flight-booking/internal/domain"
	"github.com/ambiyansyah-risyal/flight-booking/internal/usecase"
	"github.com/jmoiron/sqlx"
)

type fakeBookingRepoCLI struct {
	items  map[string]domain.Booking
	counts map[int64]int
	nextID int64
}

func newFakeBookingRepoCLI() *fakeBookingRepoCLI {
	return &fakeBookingRepoCLI{items: make(map[string]domain.Booking), counts: make(map[int64]int), nextID: 1}
}

func (f *fakeBookingRepoCLI) Create(ctx context.Context, b *domain.Booking) error {
	if _, exists := f.items[b.Reference]; exists {
		return domain.ErrBookingExists
	}
	b.ID = f.nextID
	f.nextID++
	f.items[b.Reference] = *b
	f.counts[b.ScheduleID]++
	return nil
}

func (f *fakeBookingRepoCLI) CountBySchedule(ctx context.Context, scheduleID int64) (int, error) {
	return f.counts[scheduleID], nil
}

func (f *fakeBookingRepoCLI) ListBySchedule(ctx context.Context, scheduleID int64, limit, offset int) ([]domain.Booking, error) {
	var out []domain.Booking
	for _, b := range f.items {
		if b.ScheduleID == scheduleID {
			out = append(out, b)
		}
	}
	return out, nil
}

func (f *fakeBookingRepoCLI) GetByReference(ctx context.Context, reference string) (*domain.Booking, error) {
	if b, ok := f.items[reference]; ok {
		copy := b
		return &copy, nil
	}
	return nil, domain.ErrBookingNotFound
}

type fakeBookingScheduleRepoCLI struct {
	items map[int64]domain.FlightSchedule
}

func (f *fakeBookingScheduleRepoCLI) Create(ctx context.Context, s *domain.FlightSchedule) error {
	return nil
}

func (f *fakeBookingScheduleRepoCLI) GetByID(ctx context.Context, id int64) (*domain.FlightSchedule, error) {
	if s, ok := f.items[id]; ok {
		copy := s
		return &copy, nil
	}
	return nil, domain.ErrScheduleNotFound
}

func (f *fakeBookingScheduleRepoCLI) List(ctx context.Context, routeCode string, limit, offset int) ([]domain.FlightSchedule, error) {
	var out []domain.FlightSchedule
	for _, s := range f.items {
		if routeCode == "" || s.RouteCode == routeCode {
			out = append(out, s)
		}
	}
	return out, nil
}

func (f *fakeBookingScheduleRepoCLI) Delete(ctx context.Context, id int64) error { return nil }

type fakeRouteRepoBookingCLI struct {
	items []domain.Route
}

func (f *fakeRouteRepoBookingCLI) Create(ctx context.Context, r *domain.Route) error { return nil }

func (f *fakeRouteRepoBookingCLI) GetByCode(ctx context.Context, code string) (*domain.Route, error) {
	for _, r := range f.items {
		if r.Code == code {
			copy := r
			return &copy, nil
		}
	}
	return nil, domain.ErrRouteNotFound
}

func (f *fakeRouteRepoBookingCLI) List(ctx context.Context, limit, offset int) ([]domain.Route, error) {
	return f.items, nil
}

func (f *fakeRouteRepoBookingCLI) Delete(ctx context.Context, code string) error { return nil }

type fakeAirplaneRepoBookingCLI struct {
	items map[string]domain.Airplane
}

func newFakeAirplaneRepoBookingCLI() *fakeAirplaneRepoBookingCLI {
	return &fakeAirplaneRepoBookingCLI{items: make(map[string]domain.Airplane)}
}

func (f *fakeAirplaneRepoBookingCLI) Create(ctx context.Context, a *domain.Airplane) error {
	return nil
}

func (f *fakeAirplaneRepoBookingCLI) GetByCode(ctx context.Context, code string) (*domain.Airplane, error) {
	if a, ok := f.items[code]; ok {
		copy := a
		return &copy, nil
	}
	return nil, domain.ErrAirplaneNotFound
}

func (f *fakeAirplaneRepoBookingCLI) List(ctx context.Context, limit, offset int) ([]domain.Airplane, error) {
	var out []domain.Airplane
	for _, a := range f.items {
		out = append(out, a)
	}
	return out, nil
}

func (f *fakeAirplaneRepoBookingCLI) UpdateSeats(ctx context.Context, code string, seats int) error {
	return nil
}

func (f *fakeAirplaneRepoBookingCLI) Delete(ctx context.Context, code string) error { return nil }

func TestBookingCLI_Flow(t *testing.T) {
	oldDB, oldBookingRepo, oldScheduleRepo, oldRouteRepo, oldAirplaneRepo := newBookingDB, newBookingRepo, newBookingScheduleRepo, newBookingRouteRepo, newBookingAirplaneRepo
	t.Cleanup(func() {
		newBookingDB = oldDB
		newBookingRepo = oldBookingRepo
		newBookingScheduleRepo = oldScheduleRepo
		newBookingRouteRepo = oldRouteRepo
		newBookingAirplaneRepo = oldAirplaneRepo
	})

	newBookingDB = func(string) (*sqlx.DB, error) {
		db, _, err := sqlmock.New()
		if err != nil {
			return nil, fmt.Errorf("sqlmock: %w", err)
		}
		return sqlx.NewDb(db, "pgx"), nil
	}

	bookings := newFakeBookingRepoCLI()
	schedules := &fakeBookingScheduleRepoCLI{items: map[int64]domain.FlightSchedule{1: {ID: 1, RouteCode: "RT1", AirplaneCode: "A320", DepartureDate: "2025-01-01"}}}
	routes := &fakeRouteRepoBookingCLI{items: []domain.Route{{Code: "RT1", OriginCode: "CGK", DestinationCode: "SIN"}}}
	airplanes := newFakeAirplaneRepoBookingCLI()
	airplanes.items["A320"] = domain.Airplane{Code: "A320", SeatCapacity: 2}

	newBookingRepo = func(*sqlx.DB) domain.BookingRepository { return bookings }
	newBookingScheduleRepo = func(*sqlx.DB) domain.FlightScheduleRepository { return schedules }
	newBookingRouteRepo = func(*sqlx.DB) domain.RouteRepository { return routes }
	newBookingAirplaneRepo = func(*sqlx.DB) domain.AirplaneRepository { return airplanes }

	t.Setenv("FLIGHT_DB_HOST", "localhost")

	os.Args = []string{"flight-booking", "booking", "search", "--origin", "CGK", "--destination", "SIN"}
	if err := Execute(); err != nil {
		t.Fatalf("search: %v", err)
	}

	os.Args = []string{"flight-booking", "booking", "book", "--schedule", "1", "--name", "Alice"}
	if err := Execute(); err != nil {
		t.Fatalf("book: %v", err)
	}

	os.Args = []string{"flight-booking", "booking", "list", "--schedule", "1"}
	if err := Execute(); err != nil {
		t.Fatalf("list: %v", err)
	}

	var ref string
	for k := range bookings.items {
		ref = k
	}
	if ref == "" {
		t.Fatalf("expected booking reference after create")
	}

	os.Args = []string{"flight-booking", "booking", "get", ref}
	if err := Execute(); err != nil {
		t.Fatalf("get: %v", err)
	}
}

func TestBookingCLI_MissingFlags(t *testing.T) {
	t.Setenv("FLIGHT_DB_HOST", "localhost")
	os.Args = []string{"flight-booking", "booking", "book"}
	if err := Execute(); err == nil {
		t.Fatalf("expected error for missing flags")
	}
}

// TestWriteTransitFlightOptions tests the WriteTransitFlightOptions method
func TestWriteTransitFlightOptions(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	t.Cleanup(func() {
		os.Stdout = old
	})

	writer := &RealOutputWriter{}
	options := []usecase.TransitOption{
		{
			FirstLeg: usecase.FlightOption{
				ScheduleID:      1,
				RouteCode:       "RT1",
				OriginCode:      "CGK",
				DestinationCode: "SIN",
				AirplaneCode:    "A320",
				DepartureDate:   "2025-01-01",
				SeatsAvailable:  100,
				TotalSeats:      150,
			},
			Intermediate: "KUL",
			SecondLeg: usecase.FlightOption{
				ScheduleID:      2,
				RouteCode:       "RT2",
				OriginCode:      "SIN",
				DestinationCode: "BKK",
				AirplaneCode:    "B737",
				DepartureDate:   "2025-01-01",
				SeatsAvailable:  80,
				TotalSeats:      120,
			},
			TotalAvailable: 5,
		},
	}

	err := writer.WriteTransitFlightOptions(options)
	if err != nil {
		t.Errorf("WriteTransitFlightOptions returned error: %v", err)
	}

	// Flush the pipe
	w.Close()
	_, _ = r.Read(make([]byte, 1024)) // Read the output to prevent blocking
}

// TestWriteNoTransitMessage tests the WriteNoTransitMessage method
func TestWriteNoTransitMessage(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	t.Cleanup(func() {
		os.Stdout = old
	})

	writer := &RealOutputWriter{}
	writer.WriteNoTransitMessage()

	// Flush the pipe
	w.Close()
	_, _ = r.Read(make([]byte, 1024)) // Read the output to prevent blocking
}
