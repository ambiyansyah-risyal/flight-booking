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

// TransitOption captures a connecting flight route with remaining seat inventory on each leg.
type TransitOption struct {
	FirstLeg       FlightOption
	SecondLeg      FlightOption
	Intermediate   string // Intermediate airport code
	TotalAvailable int    // Limited by the leg with fewer seats
}

// SearchTransitFlights finds connecting schedules between two airports via intermediate airports with available seats on both legs.
func (u *BookingUsecase) SearchTransitFlights(ctx context.Context, originCode, destinationCode, departureDate string) ([]TransitOption, error) {
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

	// First, get all routes that start from the origin
	originRoutes, err := u.routes.List(ctx, 500, 0)
	if err != nil {
		return nil, err
	}

	var validTransitOptions []TransitOption

	// Find intermediate airports - airports reachable from origin but not the destination
	var intermediateAirports []string
	for _, route := range originRoutes {
		if route.OriginCode == origin && route.DestinationCode != destination {
			// Check if there are schedules for this route on the specified date
			schedules, err := u.schedules.List(ctx, route.Code, 500, 0)
			if err != nil {
				continue // Skip this route if there's an error
			}
			
			hasScheduleOnDate := false
			for _, sched := range schedules {
				if date == "" || sched.DepartureDate == date {
					hasScheduleOnDate = true
					break
				}
			}
			
			if hasScheduleOnDate {
				// Add the destination of this route as a potential intermediate airport
				intermediateAirports = append(intermediateAirports, route.DestinationCode)
			}
		}
	}

	// For each potential intermediate airport, check if there's a route from there to the destination
	for _, intermediate := range intermediateAirports {
		// Get the first leg (origin to intermediate)
		firstLegRoutes := []domain.Route{}
		for _, route := range originRoutes {
			if route.OriginCode == origin && route.DestinationCode == intermediate {
				firstLegRoutes = append(firstLegRoutes, route)
			}
		}
		
		if len(firstLegRoutes) == 0 {
			continue
		}
		
		// Get the second leg (intermediate to destination)
		secondLegRoutes := []domain.Route{}
		for _, route := range originRoutes {
			if route.OriginCode == intermediate && route.DestinationCode == destination {
				secondLegRoutes = append(secondLegRoutes, route)
			}
		}
		
		if len(secondLegRoutes) == 0 {
			continue
		}

		// Find valid schedules for the first leg
		for _, firstRoute := range firstLegRoutes {
			firstSchedules, err := u.schedules.List(ctx, firstRoute.Code, 500, 0)
			if err != nil {
				continue
			}

			for _, firstSched := range firstSchedules {
				if date != "" && firstSched.DepartureDate != date {
					continue
				}

				// Get available seats for first leg
				firstPlane, err := u.airplanes.GetByCode(ctx, firstSched.AirplaneCode)
				if err != nil {
					continue
				}
				firstBooked, err := u.bookings.CountBySchedule(ctx, firstSched.ID)
				if err != nil {
					continue
				}
				firstAvailable := firstPlane.SeatCapacity - firstBooked

				if firstAvailable <= 0 {
					continue
				}

				// Now find valid schedules for the second leg
				for _, secondRoute := range secondLegRoutes {
					secondSchedules, err := u.schedules.List(ctx, secondRoute.Code, 500, 0)
					if err != nil {
						continue
					}

					for _, secondSched := range secondSchedules {
						// In a real system, you'd check that the second leg departs after the first leg arrives
						// For now, we just ensure the second leg is on the requested date if specified
						if date != "" && secondSched.DepartureDate != date {
							continue
						}

						// Get available seats for second leg
						secondPlane, err := u.airplanes.GetByCode(ctx, secondSched.AirplaneCode)
						if err != nil {
							continue
						}
						secondBooked, err := u.bookings.CountBySchedule(ctx, secondSched.ID)
						if err != nil {
							continue
						}
						secondAvailable := secondPlane.SeatCapacity - secondBooked

						if secondAvailable <= 0 {
							continue
						}

						// The total available seats is limited by the leg with fewer seats
						totalAvailable := firstAvailable
						if secondAvailable < totalAvailable {
							totalAvailable = secondAvailable
						}

						if totalAvailable > 0 {
							transitOption := TransitOption{
								FirstLeg: FlightOption{
									ScheduleID:      firstSched.ID,
									RouteCode:       firstRoute.Code,
									OriginCode:      firstRoute.OriginCode,
									DestinationCode: firstRoute.DestinationCode,
									AirplaneCode:    firstSched.AirplaneCode,
									DepartureDate:   firstSched.DepartureDate,
									SeatsAvailable:  firstAvailable,
									TotalSeats:      firstPlane.SeatCapacity,
								},
								SecondLeg: FlightOption{
									ScheduleID:      secondSched.ID,
									RouteCode:       secondRoute.Code,
									OriginCode:      secondRoute.OriginCode,
									DestinationCode: secondRoute.DestinationCode,
									AirplaneCode:    secondSched.AirplaneCode,
									DepartureDate:   secondSched.DepartureDate,
									SeatsAvailable:  secondAvailable,
									TotalSeats:      secondPlane.SeatCapacity,
								},
								Intermediate:   intermediate,
								TotalAvailable: totalAvailable,
							}
							validTransitOptions = append(validTransitOptions, transitOption)
						}
					}
				}
			}
		}
	}

	return validTransitOptions, nil
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
