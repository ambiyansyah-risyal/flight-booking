package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	sqlxrepo "github.com/ambiyansyah-risyal/flight-booking/internal/adapter/repository/sqlx"
	"github.com/ambiyansyah-risyal/flight-booking/internal/config"
	"github.com/ambiyansyah-risyal/flight-booking/internal/domain"
	"github.com/ambiyansyah-risyal/flight-booking/internal/usecase"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
)

func newBookingCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "booking", Short: "Manage direct flight bookings"}
	cmd.AddCommand(newBookingSearchCmd())
	cmd.AddCommand(newBookingCreateCmd())
	cmd.AddCommand(newBookingGetCmd())
	cmd.AddCommand(newBookingListCmd())
	return cmd
}

var (
	newBookingDB           = func(dsn string) (*sqlx.DB, error) { return sqlxrepo.New(dsn) }
	newBookingRepo         = func(db *sqlx.DB) domain.BookingRepository { return sqlxrepo.NewBookingRepository(db) }
	newBookingScheduleRepo = func(db *sqlx.DB) domain.FlightScheduleRepository { return sqlxrepo.NewScheduleRepository(db) }
	newBookingRouteRepo    = func(db *sqlx.DB) domain.RouteRepository { return sqlxrepo.NewRouteRepository(db) }
	newBookingAirplaneRepo = func(db *sqlx.DB) domain.AirplaneRepository { return sqlxrepo.NewAirplaneRepository(db) }
)

func withBookingUsecase(run func(*usecase.BookingUsecase) error) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	db, err := newBookingDB(cfg.Database.DSN())
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	uc := usecase.NewBookingUsecase(newBookingRepo(db), newBookingScheduleRepo(db), newBookingRouteRepo(db), newBookingAirplaneRepo(db))
	return run(uc)
}

func newBookingSearchCmd() *cobra.Command {
	var origin, destination, departure string
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search direct flights with available seats",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withBookingUsecase(func(uc *usecase.BookingUsecase) error {
				options, err := uc.SearchDirectFlights(context.Background(), origin, destination, departure)
				if err != nil {
					return err
				}
				tw := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
				_, _ = fmt.Fprintln(tw, "SCHEDULE\tROUTE\tDATE\tAIRPLANE\tSEATS LEFT\tTOTAL SEATS")
				for _, opt := range options {
					_, _ = fmt.Fprintf(tw, "%d\t%s->%s\t%s\t%s\t%d\t%d\n", opt.ScheduleID, opt.OriginCode, opt.DestinationCode, opt.DepartureDate, opt.AirplaneCode, opt.SeatsAvailable, opt.TotalSeats)
				}
				return tw.Flush()
			})
		},
	}
	cmd.Flags().StringVar(&origin, "origin", "", "origin airport code")
	cmd.Flags().StringVar(&destination, "destination", "", "destination airport code")
	cmd.Flags().StringVar(&departure, "date", "", "optional departure date (YYYY-MM-DD)")
	_ = cmd.MarkFlagRequired("origin")
	_ = cmd.MarkFlagRequired("destination")
	return cmd
}

func newBookingCreateCmd() *cobra.Command {
	var scheduleID int64
	var passenger string
	cmd := &cobra.Command{
		Use:   "book",
		Short: "Create a new booking for a schedule",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withBookingUsecase(func(uc *usecase.BookingUsecase) error {
				booking, err := uc.Create(context.Background(), scheduleID, passenger)
				if err != nil {
					return err
				}
				fmt.Printf("booking confirmed: %s seat %d\n", booking.Reference, booking.SeatNumber)
				return nil
			})
		},
	}
	cmd.Flags().Int64Var(&scheduleID, "schedule", 0, "schedule identifier")
	cmd.Flags().StringVar(&passenger, "name", "", "passenger full name")
	_ = cmd.MarkFlagRequired("schedule")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newBookingGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <reference>",
		Short: "Retrieve a booking by its reference",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reference := args[0]
			return withBookingUsecase(func(uc *usecase.BookingUsecase) error {
				booking, err := uc.GetByReference(context.Background(), reference)
				if err != nil {
					return err
				}
				fmt.Printf("reference: %s\npassenger: %s\nschedule: %d\nseat: %d\nstatus: %s\n", booking.Reference, booking.PassengerName, booking.ScheduleID, booking.SeatNumber, booking.Status)
				return nil
			})
		},
	}
	return cmd
}

func newBookingListCmd() *cobra.Command {
	var scheduleID int64
	var limit, offset int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List bookings for a schedule",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withBookingUsecase(func(uc *usecase.BookingUsecase) error {
				bookings, err := uc.ListBySchedule(context.Background(), scheduleID, limit, offset)
				if err != nil {
					return err
				}
				tw := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
				_, _ = fmt.Fprintln(tw, "REFERENCE\tPASSENGER\tSEAT\tSTATUS\tCREATED")
				for _, b := range bookings {
					_, _ = fmt.Fprintf(tw, "%s\t%s\t%d\t%s\t%s\n", b.Reference, b.PassengerName, b.SeatNumber, b.Status, b.CreatedAt)
				}
				return tw.Flush()
			})
		},
	}
	cmd.Flags().Int64Var(&scheduleID, "schedule", 0, "schedule identifier")
	cmd.Flags().IntVar(&limit, "limit", 50, "max bookings to list")
	cmd.Flags().IntVar(&offset, "offset", 0, "items to skip")
	_ = cmd.MarkFlagRequired("schedule")
	return cmd
}
