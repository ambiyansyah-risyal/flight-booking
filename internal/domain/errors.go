package domain

import "errors"

var (
    ErrInvalidAirportCode = errors.New("invalid airport code")
    ErrInvalidAirportCity = errors.New("invalid airport city")
    ErrAirportExists      = errors.New("airport already exists")
    ErrAirportNotFound    = errors.New("airport not found")
)

