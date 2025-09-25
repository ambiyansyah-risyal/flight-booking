package domain

import "testing"

func TestBookingValidate(t *testing.T) {
	b := Booking{
		Reference:     "bk-123456",
		ScheduleID:    10,
		PassengerName: "John Doe",
		SeatNumber:    1,
		Status:        BookingStatusConfirmed,
	}
	b.Normalize()
	if err := b.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
}

func TestBookingValidateErrors(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Booking)
		want   error
	}{
		{"schedule", func(b *Booking) { b.ScheduleID = 0 }, ErrInvalidScheduleID},
		{"passenger", func(b *Booking) { b.PassengerName = "" }, ErrInvalidPassengerName},
		{"reference", func(b *Booking) { b.Reference = "ab" }, ErrInvalidBookingReference},
		{"seat", func(b *Booking) { b.SeatNumber = 0 }, ErrInvalidSeatNumber},
		{"status", func(b *Booking) { b.Status = "unknown" }, ErrInvalidBookingStatus},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := Booking{
				Reference:     "bk-123456",
				ScheduleID:    10,
				PassengerName: "Jane",
				SeatNumber:    1,
				Status:        BookingStatusConfirmed,
			}
			tc.mutate(&b)
			b.Normalize()
			if err := b.Validate(); err != tc.want {
				t.Fatalf("want %v, got %v", tc.want, err)
			}
		})
	}
}
