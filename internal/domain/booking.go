package domain

import "strings"

const (
	BookingStatusConfirmed = "CONFIRMED"
	BookingStatusCancelled = "CANCELLED"
)

// Booking represents a confirmed seat for a passenger on a scheduled flight.
type Booking struct {
	ID            int64
	Reference     string
	ScheduleID    int64
	PassengerName string
	SeatNumber    int
	Status        string
	CreatedAt     string
}

// Normalize trims and uppercases key booking fields for consistency.
func (b *Booking) Normalize() {
	b.Reference = strings.ToUpper(strings.TrimSpace(b.Reference))
	b.PassengerName = strings.TrimSpace(b.PassengerName)
	b.Status = strings.ToUpper(strings.TrimSpace(b.Status))
}

// Validate ensures the booking is structurally sound before persistence.
func (b Booking) Validate() error {
	if b.ScheduleID <= 0 {
		return ErrInvalidScheduleID
	}
	if name := strings.TrimSpace(b.PassengerName); len(name) == 0 || len(name) > 128 {
		return ErrInvalidPassengerName
	}
	if ref := strings.TrimSpace(b.Reference); len(ref) < 6 || len(ref) > 32 {
		return ErrInvalidBookingReference
	}
	if b.SeatNumber <= 0 {
		return ErrInvalidSeatNumber
	}
	switch b.Status {
	case BookingStatusConfirmed, BookingStatusCancelled:
		// ok
	default:
		return ErrInvalidBookingStatus
	}
	return nil
}
