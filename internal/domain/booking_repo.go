package domain

import "context"

// BookingRepository defines persistence operations for flight bookings.
type BookingRepository interface {
	Create(ctx context.Context, b *Booking) error
	CountBySchedule(ctx context.Context, scheduleID int64) (int, error)
	ListBySchedule(ctx context.Context, scheduleID int64, limit, offset int) ([]Booking, error)
	GetByReference(ctx context.Context, reference string) (*Booking, error)
}
