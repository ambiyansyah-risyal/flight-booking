package domain

import "testing"

func TestAirportNormalize(t *testing.T) {
    a := &Airport{Code: "  cgk ", City: " Jakarta  "}
    a.Normalize()
    if a.Code != "CGK" {
        t.Fatalf("expected code CGK, got %q", a.Code)
    }
    if a.City != "Jakarta" {
        t.Fatalf("expected city Jakarta, got %q", a.City)
    }
}

func TestAirportValidate_OK(t *testing.T) {
    a := Airport{Code: "CGK", City: "Jakarta"}
    if err := a.Validate(); err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}

func TestAirportValidate_Errors(t *testing.T) {
    cases := []struct{
        name string
        in   Airport
        want error
    }{
        {"empty code", Airport{Code: "", City: "City"}, ErrInvalidAirportCode},
        {"long code", Airport{Code: "ABCDEFGHI", City: "City"}, ErrInvalidAirportCode},
        {"empty city", Airport{Code: "CGK", City: ""}, ErrInvalidAirportCity},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            if err := tc.in.Validate(); err != tc.want {
                t.Fatalf("want %v, got %v", tc.want, err)
            }
        })
    }
}

