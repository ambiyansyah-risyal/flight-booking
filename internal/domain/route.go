package domain

import "strings"

// Route represents a direct path between two airports.
type Route struct {
	ID              int64
	Code            string
	OriginCode      string
	DestinationCode string
	CreatedAt       string
}

// Normalize trims whitespace and uppercases route identifiers and airport codes.
func (r *Route) Normalize() {
	r.Code = strings.ToUpper(strings.TrimSpace(r.Code))
	r.OriginCode = strings.ToUpper(strings.TrimSpace(r.OriginCode))
	r.DestinationCode = strings.ToUpper(strings.TrimSpace(r.DestinationCode))
}

// Validate ensures the route has a code and distinct origin/destination airports.
func (r Route) Validate() error {
	if len(strings.TrimSpace(r.Code)) == 0 || len(r.Code) > 16 {
		return ErrInvalidRouteCode
	}
	if len(strings.TrimSpace(r.OriginCode)) == 0 || len(r.OriginCode) > 8 {
		return ErrInvalidRouteAirports
	}
	if len(strings.TrimSpace(r.DestinationCode)) == 0 || len(r.DestinationCode) > 8 {
		return ErrInvalidRouteAirports
	}
	if r.OriginCode == r.DestinationCode {
		return ErrInvalidRouteAirports
	}
	return nil
}
