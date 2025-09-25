package domain

import "testing"

func TestRouteNormalizeValidate(t *testing.T) {
	r := &Route{Code: " rt-01 ", OriginCode: " cgk ", DestinationCode: " dps "}
	r.Normalize()
	if r.Code != "RT-01" || r.OriginCode != "CGK" || r.DestinationCode != "DPS" {
		t.Fatalf("normalize failed: %+v", r)
	}
	if err := r.Validate(); err != nil {
		t.Fatalf("expected valid route, got %v", err)
	}
}

func TestRouteValidate_Errors(t *testing.T) {
	cases := []struct {
		r   Route
		err error
	}{
		{Route{Code: "", OriginCode: "CGK", DestinationCode: "DPS"}, ErrInvalidRouteCode},
		{Route{Code: "R1", OriginCode: "", DestinationCode: "DPS"}, ErrInvalidRouteAirports},
		{Route{Code: "R1", OriginCode: "CGK", DestinationCode: ""}, ErrInvalidRouteAirports},
		{Route{Code: "R1", OriginCode: "CGK", DestinationCode: "CGK"}, ErrInvalidRouteAirports},
	}
	for _, tc := range cases {
		if err := tc.r.Validate(); err != tc.err {
			t.Fatalf("want %v, got %v", tc.err, err)
		}
	}
}
