package domain

import "testing"

func TestFlightScheduleNormalizeValidate(t *testing.T) {
	s := &FlightSchedule{RouteCode: " rt-01 ", AirplaneCode: " a320 ", DepartureDate: "2025-01-02"}
	s.Normalize()
	if s.RouteCode != "RT-01" || s.AirplaneCode != "A320" {
		t.Fatalf("normalize failed: %+v", s)
	}
	if err := s.Validate(); err != nil {
		t.Fatalf("expected valid schedule, got %v", err)
	}
}

func TestFlightScheduleValidate_Errors(t *testing.T) {
	cases := []struct {
		s   FlightSchedule
		err error
	}{
		{FlightSchedule{RouteCode: "", AirplaneCode: "A320", DepartureDate: "2025-01-01"}, ErrInvalidScheduleRoute},
		{FlightSchedule{RouteCode: "R1", AirplaneCode: "", DepartureDate: "2025-01-01"}, ErrInvalidScheduleAirplane},
		{FlightSchedule{RouteCode: "R1", AirplaneCode: "A320", DepartureDate: "bad-date"}, ErrInvalidScheduleDate},
	}
	for _, tc := range cases {
		if err := tc.s.Validate(); err != tc.err {
			t.Fatalf("want %v, got %v", tc.err, err)
		}
	}
}
