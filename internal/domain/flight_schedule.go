package domain

import (
	"strings"
	"time"
)

// FlightSchedule represents a planned flight on a specific date for a given route and airplane.
type FlightSchedule struct {
	ID            int64
	RouteCode     string
	AirplaneCode  string
	DepartureDate string // YYYY-MM-DD in UTC
	CreatedAt     string
}

// Normalize trims and uppercases codes to keep consistency across adapters.
func (s *FlightSchedule) Normalize() {
	s.RouteCode = strings.ToUpper(strings.TrimSpace(s.RouteCode))
	s.AirplaneCode = strings.ToUpper(strings.TrimSpace(s.AirplaneCode))
	s.DepartureDate = strings.TrimSpace(s.DepartureDate)
}

// Validate checks that codes are present and the departure date is a valid ISO date.
func (s FlightSchedule) Validate() error {
	if len(strings.TrimSpace(s.RouteCode)) == 0 || len(s.RouteCode) > 16 {
		return ErrInvalidScheduleRoute
	}
	if len(strings.TrimSpace(s.AirplaneCode)) == 0 || len(s.AirplaneCode) > 16 {
		return ErrInvalidScheduleAirplane
	}
	if _, err := time.Parse("2006-01-02", s.DepartureDate); err != nil {
		return ErrInvalidScheduleDate
	}
	return nil
}
