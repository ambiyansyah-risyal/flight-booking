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

func newRouteCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "route", Short: "Manage flight routes"}
	cmd.AddCommand(newRouteCreateCmd())
	cmd.AddCommand(newRouteListCmd())
	cmd.AddCommand(newRouteDeleteCmd())
	return cmd
}

var (
	newRouteDB          = func(dsn string) (*sqlx.DB, error) { return sqlxrepo.New(dsn) }
	newRouteRepo        = func(db *sqlx.DB) domain.RouteRepository { return sqlxrepo.NewRouteRepository(db) }
	newRouteAirportRepo = func(db *sqlx.DB) domain.AirportRepository { return sqlxrepo.NewAirportRepository(db) }
)

func withRouteUsecase(run func(*usecase.RouteUsecase) error) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	db, err := newRouteDB(cfg.Database.DSN())
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	uc := usecase.NewRouteUsecase(newRouteRepo(db), newRouteAirportRepo(db))
	return run(uc)
}

func newRouteCreateCmd() *cobra.Command {
	var code, origin, destination string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new route",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withRouteUsecase(func(uc *usecase.RouteUsecase) error {
				route, err := uc.Create(context.Background(), code, origin, destination)
				if err != nil {
					return err
				}
				fmt.Printf("created route %s (%s -> %s)\n", route.Code, route.OriginCode, route.DestinationCode)
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&code, "code", "", "route code identifier")
	cmd.Flags().StringVar(&origin, "origin", "", "origin airport code")
	cmd.Flags().StringVar(&destination, "destination", "", "destination airport code")
	_ = cmd.MarkFlagRequired("code")
	_ = cmd.MarkFlagRequired("origin")
	_ = cmd.MarkFlagRequired("destination")
	return cmd
}

func newRouteListCmd() *cobra.Command {
	var limit, offset int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List routes",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withRouteUsecase(func(uc *usecase.RouteUsecase) error {
				routes, err := uc.List(context.Background(), limit, offset)
				if err != nil {
					return err
				}
				tw := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
				_, _ = fmt.Fprintln(tw, "CODE\tORIGIN\tDESTINATION")
				for _, r := range routes {
					_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\n", r.Code, r.OriginCode, r.DestinationCode)
				}
				return tw.Flush()
			})
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 50, "max items to list")
	cmd.Flags().IntVar(&offset, "offset", 0, "items to skip")
	return cmd
}

func newRouteDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <code>",
		Short: "Delete a route by code",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			code := args[0]
			return withRouteUsecase(func(uc *usecase.RouteUsecase) error {
				if err := uc.Delete(context.Background(), code); err != nil {
					return err
				}
				fmt.Printf("deleted route %s\n", code)
				return nil
			})
		},
	}
	return cmd
}
