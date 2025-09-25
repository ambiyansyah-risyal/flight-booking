package usecase

import (
	"context"
	"strings"
	"testing"

	"github.com/ambiyansyah-risyal/flight-booking/internal/domain"
)

type fakeBookingRepo struct {
	items     map[string]domain.Booking
	counts    map[int64]int
	createErr error
}

func newFakeBookingRepo() *fakeBookingRepo {
	return &fakeBookingRepo{items: make(map[string]domain.Booking), counts: make(map[int64]int)}
}

func (f *fakeBookingRepo) Create(ctx context.Context, b *domain.Booking) error {
	if f.createErr != nil {
		return f.createErr
	}
	if _, ok := f.items[b.Reference]; ok {
		return domain.ErrBookingExists
	}
	b.ID = int64(len(f.items) + 1)
	f.items[b.Reference] = *b
	f.counts[b.ScheduleID]++
	return nil
}

func (f *fakeBookingRepo) CountBySchedule(ctx context.Context, scheduleID int64) (int, error) {
	return f.counts[scheduleID], nil
}

func (f *fakeBookingRepo) ListBySchedule(ctx context.Context, scheduleID int64, limit, offset int) ([]domain.Booking, error) {
	var out []domain.Booking
	for _, b := range f.items {
		if b.ScheduleID == scheduleID {
			out = append(out, b)
		}
	}
	return out, nil
}

func (f *fakeBookingRepo) GetByReference(ctx context.Context, reference string) (*domain.Booking, error) {
	if b, ok := f.items[reference]; ok {
		copy := b
		return &copy, nil
	}
	return nil, domain.ErrBookingNotFound
}

type fakeScheduleRepoBooking struct {
	items map[int64]domain.FlightSchedule
}

func newFakeScheduleRepoBooking() *fakeScheduleRepoBooking {
	return &fakeScheduleRepoBooking{items: make(map[int64]domain.FlightSchedule)}
}

func (f *fakeScheduleRepoBooking) Create(ctx context.Context, s *domain.FlightSchedule) error {
	f.items[s.ID] = *s
	return nil
}

func (f *fakeScheduleRepoBooking) GetByID(ctx context.Context, id int64) (*domain.FlightSchedule, error) {
	if s, ok := f.items[id]; ok {
		copy := s
		return &copy, nil
	}
	return nil, domain.ErrScheduleNotFound
}

func (f *fakeScheduleRepoBooking) List(ctx context.Context, routeCode string, limit, offset int) ([]domain.FlightSchedule, error) {
	var out []domain.FlightSchedule
	for _, s := range f.items {
		if routeCode == "" || s.RouteCode == routeCode {
			out = append(out, s)
		}
	}
	return out, nil
}

func (f *fakeScheduleRepoBooking) Delete(ctx context.Context, id int64) error { return nil }

type fakeRouteRepoBooking struct {
	items []domain.Route
}

func (f *fakeRouteRepoBooking) Create(ctx context.Context, r *domain.Route) error { return nil }

func (f *fakeRouteRepoBooking) GetByCode(ctx context.Context, code string) (*domain.Route, error) {
	for _, r := range f.items {
		if r.Code == code {
			copy := r
			return &copy, nil
		}
	}
	return nil, domain.ErrRouteNotFound
}

func (f *fakeRouteRepoBooking) List(ctx context.Context, limit, offset int) ([]domain.Route, error) {
	return f.items, nil
}

func (f *fakeRouteRepoBooking) Delete(ctx context.Context, code string) error { return nil }

type fakeAirplaneRepoBooking struct {
	items map[string]domain.Airplane
}

func newFakeAirplaneRepoBooking() *fakeAirplaneRepoBooking {
	return &fakeAirplaneRepoBooking{items: make(map[string]domain.Airplane)}
}

func (f *fakeAirplaneRepoBooking) Create(ctx context.Context, a *domain.Airplane) error { return nil }

func (f *fakeAirplaneRepoBooking) GetByCode(ctx context.Context, code string) (*domain.Airplane, error) {
	if a, ok := f.items[code]; ok {
		copy := a
		return &copy, nil
	}
	return nil, domain.ErrAirplaneNotFound
}

func (f *fakeAirplaneRepoBooking) List(ctx context.Context, limit, offset int) ([]domain.Airplane, error) {
	var out []domain.Airplane
	for _, a := range f.items {
		out = append(out, a)
	}
	return out, nil
}

func (f *fakeAirplaneRepoBooking) UpdateSeats(ctx context.Context, code string, seats int) error {
	return nil
}

func (f *fakeAirplaneRepoBooking) Delete(ctx context.Context, code string) error { return nil }

func TestBookingUsecase_CreateAndSeatAllocation(t *testing.T) {
	bookings := newFakeBookingRepo()
	schedules := newFakeScheduleRepoBooking()
	schedules.items[1] = domain.FlightSchedule{ID: 1, RouteCode: "RT1", AirplaneCode: "A320", DepartureDate: "2025-01-01"}
	airplanes := newFakeAirplaneRepoBooking()
	airplanes.items["A320"] = domain.Airplane{Code: "A320", SeatCapacity: 2}
	routes := &fakeRouteRepoBooking{}

	uc := NewBookingUsecase(bookings, schedules, routes, airplanes)

	b1, err := uc.Create(context.Background(), 1, " Alice ")
	if err != nil {
		t.Fatalf("create first: %v", err)
	}
	if b1.SeatNumber != 1 {
		t.Fatalf("expected seat 1, got %d", b1.SeatNumber)
	}

	b2, err := uc.Create(context.Background(), 1, "Bob")
	if err != nil {
		t.Fatalf("create second: %v", err)
	}
	if b2.SeatNumber != 2 {
		t.Fatalf("expected seat 2, got %d", b2.SeatNumber)
	}

	if _, err := uc.Create(context.Background(), 1, "Charlie"); err != domain.ErrFlightFull {
		t.Fatalf("expected flight full, got %v", err)
	}

	if got := strings.ToUpper(b1.Reference); got != b1.Reference {
		t.Fatalf("reference not uppercased: %s", b1.Reference)
	}
}

func TestBookingUsecase_SearchDirectFlights(t *testing.T) {
	bookings := newFakeBookingRepo()
	bookings.counts[1] = 1
	bookings.counts[2] = 2

	schedules := newFakeScheduleRepoBooking()
	schedules.items[1] = domain.FlightSchedule{ID: 1, RouteCode: "RT1", AirplaneCode: "A320", DepartureDate: "2025-01-01"}
	schedules.items[2] = domain.FlightSchedule{ID: 2, RouteCode: "RT1", AirplaneCode: "A320", DepartureDate: "2025-01-02"}

	airplanes := newFakeAirplaneRepoBooking()
	airplanes.items["A320"] = domain.Airplane{Code: "A320", SeatCapacity: 2}

	routes := &fakeRouteRepoBooking{items: []domain.Route{{Code: "RT1", OriginCode: "CGK", DestinationCode: "SIN"}}}

	uc := NewBookingUsecase(bookings, schedules, routes, airplanes)

	options, err := uc.SearchDirectFlights(context.Background(), "cgk", "sin", "2025-01-01")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(options) != 1 {
		t.Fatalf("expected 1 option, got %d", len(options))
	}
	if options[0].SeatsAvailable != 1 {
		t.Fatalf("expected 1 seat left, got %d", options[0].SeatsAvailable)
	}

	options, err = uc.SearchDirectFlights(context.Background(), "CGK", "SIN", "")
	if err != nil {
		t.Fatalf("search no date: %v", err)
	}
	if len(options) != 1 {
		t.Fatalf("expected only schedules with seats, got %d", len(options))
	}

	if _, err := uc.SearchDirectFlights(context.Background(), "CGK", "SIN", "bad"); err != domain.ErrInvalidScheduleDate {
		t.Fatalf("expected invalid date, got %v", err)
	}
}

func TestBookingUsecase_GetByReference(t *testing.T) {
	bookings := newFakeBookingRepo()
	bookings.items["BK-AAAAAA"] = domain.Booking{Reference: "BK-AAAAAA", ScheduleID: 1, PassengerName: "Doe", SeatNumber: 1, Status: domain.BookingStatusConfirmed}
	bookings.counts[1] = 1

	schedules := newFakeScheduleRepoBooking()
	schedules.items[1] = domain.FlightSchedule{ID: 1, RouteCode: "RT1", AirplaneCode: "A320"}
	airplanes := newFakeAirplaneRepoBooking()
	airplanes.items["A320"] = domain.Airplane{Code: "A320", SeatCapacity: 2}
	routes := &fakeRouteRepoBooking{}

	uc := NewBookingUsecase(bookings, schedules, routes, airplanes)

	b, err := uc.GetByReference(context.Background(), "bk-aaaaaa")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if b.Reference != "BK-AAAAAA" {
		t.Fatalf("reference should be uppercased, got %s", b.Reference)
	}

	if _, err := uc.GetByReference(context.Background(), "short"); err != domain.ErrInvalidBookingReference {
		t.Fatalf("expected invalid reference, got %v", err)
	}
}

func TestBookingUsecase_ListBySchedule(t *testing.T) {
	bookings := newFakeBookingRepo()
	bookings.items["BK-AAAAAA"] = domain.Booking{Reference: "BK-AAAAAA", ScheduleID: 1, PassengerName: "Alice", SeatNumber: 1, Status: domain.BookingStatusConfirmed}
	bookings.counts[1] = 1

	schedules := newFakeScheduleRepoBooking()
	schedules.items[1] = domain.FlightSchedule{ID: 1, RouteCode: "RT1", AirplaneCode: "A320"}

	airplanes := newFakeAirplaneRepoBooking()
	airplanes.items["A320"] = domain.Airplane{Code: "A320", SeatCapacity: 2}

	routes := &fakeRouteRepoBooking{}
	uc := NewBookingUsecase(bookings, schedules, routes, airplanes)

	items, err := uc.ListBySchedule(context.Background(), 1, 0, -1)
	if err != nil || len(items) != 1 {
		t.Fatalf("list: err=%v len=%d", err, len(items))
	}

	if _, err := uc.ListBySchedule(context.Background(), 0, 10, 0); err != domain.ErrInvalidScheduleID {
		t.Fatalf("expected invalid schedule id, got %v", err)
	}

	if _, err := uc.ListBySchedule(context.Background(), 2, 10, 0); err != domain.ErrScheduleNotFound {
		t.Fatalf("expected schedule not found, got %v", err)
	}
}
