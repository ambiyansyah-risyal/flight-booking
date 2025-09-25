package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/ambiyansyah-risyal/flight-booking/internal/domain"
)

// FlightOption captures a direct flight schedule with remaining seat inventory.
type FlightOption struct {
	ScheduleID      int64
	RouteCode       string
	OriginCode      string
	DestinationCode string
	AirplaneCode    string
	DepartureDate   string
	SeatsAvailable  int
	TotalSeats      int
}

// BookingUsecase coordinates booking workflows across repositories.
type BookingUsecase struct {
	bookings    domain.BookingRepository
	schedules   domain.FlightScheduleRepository
	routes      domain.RouteRepository
	airplanes   domain.AirplaneRepository
	timeout     time.Duration
	generateRef func() string
}

// NewBookingUsecase builds a BookingUsecase with sane defaults.
func NewBookingUsecase(bookRepo domain.BookingRepository, scheduleRepo domain.FlightScheduleRepository, routeRepo domain.RouteRepository, airplaneRepo domain.AirplaneRepository) *BookingUsecase {
	return &BookingUsecase{
		bookings:    bookRepo,
		schedules:   scheduleRepo,
		routes:      routeRepo,
		airplanes:   airplaneRepo,
		timeout:     5 * time.Second,
		generateRef: defaultBookingReference,
	}
}

// SearchDirectFlights finds direct schedules between two airports with available seats.
func (u *BookingUsecase) SearchDirectFlights(ctx context.Context, originCode, destinationCode, departureDate string) ([]FlightOption, error) {
	origin := strings.ToUpper(strings.TrimSpace(originCode))
	destination := strings.ToUpper(strings.TrimSpace(destinationCode))
	date := strings.TrimSpace(departureDate)

	if len(origin) == 0 || len(origin) > 8 {
		return nil, domain.ErrInvalidRouteAirports
	}
	if len(destination) == 0 || len(destination) > 8 {
		return nil, domain.ErrInvalidRouteAirports
	}
	if origin == destination {
		return nil, domain.ErrInvalidRouteAirports
	}

	if date != "" {
		if _, err := time.Parse("2006-01-02", date); err != nil {
			return nil, domain.ErrInvalidScheduleDate
		}
	}

	ctx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()

	routes, err := u.routes.List(ctx, 500, 0)
	if err != nil {
		return nil, err
	}

	var matched []domain.Route
	for _, route := range routes {
		if route.OriginCode == origin && route.DestinationCode == destination {
			matched = append(matched, route)
		}
	}
	if len(matched) == 0 {
		return nil, domain.ErrRouteNotFound
	}

	planeCache := make(map[string]domain.Airplane)
	var options []FlightOption

	for _, route := range matched {
		schedules, err := u.schedules.List(ctx, route.Code, 500, 0)
		if err != nil {
			return nil, err
		}
		for _, sched := range schedules {
			if date != "" && sched.DepartureDate != date {
				continue
			}
			plane, ok := planeCache[sched.AirplaneCode]
			if !ok {
				ap, err := u.airplanes.GetByCode(ctx, sched.AirplaneCode)
				if err != nil {
					return nil, err
				}
				plane = *ap
				planeCache[sched.AirplaneCode] = plane
			}
			if plane.SeatCapacity <= 0 {
				continue
			}
			count, err := u.bookings.CountBySchedule(ctx, sched.ID)
			if err != nil {
				return nil, err
			}
			available := plane.SeatCapacity - count
			if available <= 0 {
				continue
			}
			options = append(options, FlightOption{
				ScheduleID:      sched.ID,
				RouteCode:       route.Code,
				OriginCode:      origin,
				DestinationCode: destination,
				AirplaneCode:    sched.AirplaneCode,
				DepartureDate:   sched.DepartureDate,
				SeatsAvailable:  available,
				TotalSeats:      plane.SeatCapacity,
			})
		}
	}

	return options, nil
}

// Create generates a booking for a passenger on a given schedule with automatic seat assignment.
func (u *BookingUsecase) Create(ctx context.Context, scheduleID int64, passengerName string) (*domain.Booking, error) {
	if scheduleID <= 0 {
		return nil, domain.ErrInvalidScheduleID
	}
	if len(strings.TrimSpace(passengerName)) == 0 {
		return nil, domain.ErrInvalidPassengerName
	}

	ctx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()

	sched, err := u.schedules.GetByID(ctx, scheduleID)
	if err != nil {
		return nil, err
	}
	plane, err := u.airplanes.GetByCode(ctx, sched.AirplaneCode)
	if err != nil {
		return nil, err
	}
	if plane.SeatCapacity <= 0 {
		return nil, domain.ErrInvalidSeatCapacity
	}
	count, err := u.bookings.CountBySchedule(ctx, scheduleID)
	if err != nil {
		return nil, err
	}
	if count >= plane.SeatCapacity {
		return nil, domain.ErrFlightFull
	}

	booking := &domain.Booking{
		Reference:     u.generateRef(),
		ScheduleID:    scheduleID,
		PassengerName: passengerName,
		SeatNumber:    count + 1,
		Status:        domain.BookingStatusConfirmed,
	}
	booking.Normalize()
	if err := booking.Validate(); err != nil {
		return nil, err
	}
	if err := u.bookings.Create(ctx, booking); err != nil {
		return nil, err
	}
	return booking, nil
}

// GetByReference fetches a previously created booking using its confirmation reference.
func (u *BookingUsecase) ListBySchedule(ctx context.Context, scheduleID int64, limit, offset int) ([]domain.Booking, error) {
	if scheduleID <= 0 {
		return nil, domain.ErrInvalidScheduleID
	}
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	ctx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()
	if _, err := u.schedules.GetByID(ctx, scheduleID); err != nil {
		return nil, err
	}
	return u.bookings.ListBySchedule(ctx, scheduleID, limit, offset)
}

func (u *BookingUsecase) GetByReference(ctx context.Context, reference string) (*domain.Booking, error) {
	ref := strings.ToUpper(strings.TrimSpace(reference))
	if len(ref) < 6 || len(ref) > 32 {
		return nil, domain.ErrInvalidBookingReference
	}
	ctx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()
	return u.bookings.GetByReference(ctx, ref)
}

func defaultBookingReference() string {
	buf := make([]byte, 5)
	if _, err := rand.Read(buf); err == nil {
		return fmt.Sprintf("BK-%s", strings.ToUpper(hex.EncodeToString(buf)))
	}
	return fmt.Sprintf("BK-%d", time.Now().UnixNano())
}
