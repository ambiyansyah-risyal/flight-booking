package domain

import "errors"

var (
    ErrInvalidAirportCode = errors.New("invalid airport code")
    ErrInvalidAirportCity = errors.New("invalid airport city")
    ErrAirportExists      = errors.New("airport already exists")
    ErrAirportNotFound    = errors.New("airport not found")
    ErrInvalidAirplaneCode   = errors.New("invalid airplane code")
    ErrInvalidSeatCapacity   = errors.New("invalid seat capacity")
    ErrAirplaneExists        = errors.New("airplane already exists")
    ErrAirplaneNotFound      = errors.New("airplane not found")
)
