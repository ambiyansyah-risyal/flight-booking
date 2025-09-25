package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	sqlxrepo "github.com/ambiyansyah-risyal/flight-booking/internal/adapter/repository/sqlx"
	"github.com/ambiyansyah-risyal/flight-booking/internal/config"
	"github.com/ambiyansyah-risyal/flight-booking/internal/domain"
	"github.com/ambiyansyah-risyal/flight-booking/internal/usecase"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
)

func newScheduleCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "schedule", Short: "Manage flight schedules"}
	cmd.AddCommand(newScheduleCreateCmd())
	cmd.AddCommand(newScheduleListCmd())
	cmd.AddCommand(newScheduleDeleteCmd())
	return cmd
}

var (
	newScheduleDB           = func(dsn string) (*sqlx.DB, error) { return sqlxrepo.New(dsn) }
	newScheduleRepo         = func(db *sqlx.DB) domain.FlightScheduleRepository { return sqlxrepo.NewScheduleRepository(db) }
	newScheduleRouteRepo    = func(db *sqlx.DB) domain.RouteRepository { return sqlxrepo.NewRouteRepository(db) }
	newScheduleAirplaneRepo = func(db *sqlx.DB) domain.AirplaneRepository { return sqlxrepo.NewAirplaneRepository(db) }
)

func withScheduleUsecase(run func(*usecase.ScheduleUsecase) error) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	db, err := newScheduleDB(cfg.Database.DSN())
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	uc := usecase.NewScheduleUsecase(newScheduleRepo(db), newScheduleRouteRepo(db), newScheduleAirplaneRepo(db))
	return run(uc)
}

func newScheduleCreateCmd() *cobra.Command {
	var routeCode, airplaneCode, departureDate string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new flight schedule",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withScheduleUsecase(func(uc *usecase.ScheduleUsecase) error {
				sched, err := uc.Create(context.Background(), routeCode, airplaneCode, departureDate)
				if err != nil {
					return err
				}
				fmt.Printf("scheduled route %s with %s on %s\n", sched.RouteCode, sched.AirplaneCode, sched.DepartureDate)
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&routeCode, "route", "", "route code")
	cmd.Flags().StringVar(&airplaneCode, "airplane", "", "airplane code")
	cmd.Flags().StringVar(&departureDate, "date", "", "departure date (YYYY-MM-DD)")
	_ = cmd.MarkFlagRequired("route")
	_ = cmd.MarkFlagRequired("airplane")
	_ = cmd.MarkFlagRequired("date")
	return cmd
}

func newScheduleListCmd() *cobra.Command {
	var routeCode string
	var limit, offset int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List flight schedules",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withScheduleUsecase(func(uc *usecase.ScheduleUsecase) error {
				items, err := uc.List(context.Background(), routeCode, limit, offset)
				if err != nil {
					return err
				}
				tw := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
				_, _ = fmt.Fprintln(tw, "ID\tROUTE\tAIRPLANE\tDEPARTURE")
				for _, s := range items {
					_, _ = fmt.Fprintf(tw, "%d\t%s\t%s\t%s\n", s.ID, s.RouteCode, s.AirplaneCode, s.DepartureDate)
				}
				return tw.Flush()
			})
		},
	}
	cmd.Flags().StringVar(&routeCode, "route", "", "optional route filter")
	cmd.Flags().IntVar(&limit, "limit", 50, "max items to list")
	cmd.Flags().IntVar(&offset, "offset", 0, "items to skip")
	return cmd
}

func newScheduleDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a scheduled flight by id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("parse id: %w", err)
			}
			return withScheduleUsecase(func(uc *usecase.ScheduleUsecase) error {
				if err := uc.Delete(context.Background(), id); err != nil {
					return err
				}
				fmt.Printf("deleted schedule %d\n", id)
				return nil
			})
		},
	}
	return cmd
}
