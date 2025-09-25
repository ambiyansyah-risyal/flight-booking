package domain

import "testing"

func TestAirplaneNormalizeAndValidate(t *testing.T) {
    a := &Airplane{Code: "  ab-123 ", SeatCapacity: 100}
    a.Normalize()
    if a.Code != "AB-123" { t.Fatalf("normalize code: %q", a.Code) }
    if err := a.Validate(); err != nil { t.Fatalf("unexpected err: %v", err) }
}

func TestAirplaneValidate_Errors(t *testing.T) {
    if err := (Airplane{Code:"", SeatCapacity:1}).Validate(); err != ErrInvalidAirplaneCode { t.Fatalf("want code err, got %v", err) }
    if err := (Airplane{Code:"OK", SeatCapacity:0}).Validate(); err != ErrInvalidSeatCapacity { t.Fatalf("want seat err, got %v", err) }
}

