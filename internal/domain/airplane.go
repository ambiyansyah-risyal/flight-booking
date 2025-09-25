package domain

import (
    "strings"
)

type Airplane struct {
    ID          int64
    Code        string
    SeatCapacity int
    CreatedAt   string
}

func (a *Airplane) Normalize() {
    a.Code = strings.ToUpper(strings.TrimSpace(a.Code))
}

func (a Airplane) Validate() error {
    if len(strings.TrimSpace(a.Code)) == 0 || len(a.Code) > 16 {
        return ErrInvalidAirplaneCode
    }
    if a.SeatCapacity <= 0 {
        return ErrInvalidSeatCapacity
    }
    return nil
}

