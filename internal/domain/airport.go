package domain

import (
    "strings"

    gv "github.com/asaskevich/govalidator"
)

type Airport struct {
    ID        int64
    Code      string
    City      string
    CreatedAt string // RFC3339, left as string for portability in domain
}

func (a *Airport) Normalize() {
    a.Code = strings.ToUpper(strings.TrimSpace(a.Code))
    a.City = strings.TrimSpace(a.City)
}

func (a Airport) Validate() error {
    if gv.IsNull(a.Code) || len(a.Code) > 8 {
        return ErrInvalidAirportCode
    }
    if gv.IsNull(a.City) {
        return ErrInvalidAirportCity
    }
    return nil
}

