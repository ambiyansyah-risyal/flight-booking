package cli

import (
    "context"
    "fmt"
    "os"
    "text/tabwriter"

    "github.com/ambiyansyah-risyal/flight-booking/internal/adapter/repository/sqlx"
    "github.com/ambiyansyah-risyal/flight-booking/internal/domain"
    "github.com/ambiyansyah-risyal/flight-booking/internal/config"
    "github.com/ambiyansyah-risyal/flight-booking/internal/usecase"
    "github.com/spf13/cobra"
    "github.com/jmoiron/sqlx"
)

func newAirportCmd() *cobra.Command {
    cmd := &cobra.Command{Use: "airport", Short: "Manage airports"}

    cmd.AddCommand(newAirportCreateCmd())
    cmd.AddCommand(newAirportListCmd())
    cmd.AddCommand(newAirportUpdateCmd())
    cmd.AddCommand(newAirportDeleteCmd())
    return cmd
}

// Injection points for testing
var (
    newDB = func(dsn string) (*sqlx.DB, error) { return sqlxrepo.New(dsn) }
    newAirportRepo = func(db *sqlx.DB) domain.AirportRepository { return sqlxrepo.NewAirportRepository(db) }
)

func withAirportUsecase(run func(u *usecase.AirportUsecase) error) error {
    cfg, err := config.Load()
    if err != nil { return err }
    db, err := newDB(cfg.Database.DSN())
    if err != nil { return err }
    defer func() { _ = db.Close() }()
    repo := newAirportRepo(db)
    uc := usecase.NewAirportUsecase(repo)
    return run(uc)
}

func newAirportCreateCmd() *cobra.Command {
    var code, city string
    cmd := &cobra.Command{
        Use:   "create",
        Short: "Create an airport",
        RunE: func(cmd *cobra.Command, args []string) error {
            return withAirportUsecase(func(u *usecase.AirportUsecase) error {
                a, err := u.Create(context.Background(), code, city)
                if err != nil { return err }
                fmt.Printf("created airport %s (%s)\n", a.Code, a.City)
                return nil
            })
        },
    }
    cmd.Flags().StringVar(&code, "code", "", "airport code (e.g., CGK)")
    cmd.Flags().StringVar(&city, "city", "", "city name")
    _ = cmd.MarkFlagRequired("code")
    _ = cmd.MarkFlagRequired("city")
    return cmd
}

func newAirportListCmd() *cobra.Command {
    var limit, offset int
    cmd := &cobra.Command{
        Use:   "list",
        Short: "List airports",
        RunE: func(cmd *cobra.Command, args []string) error {
            return withAirportUsecase(func(u *usecase.AirportUsecase) error {
                items, err := u.List(context.Background(), limit, offset)
                if err != nil { return err }
                tw := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
                _, _ = fmt.Fprintln(tw, "CODE\tCITY")
                for _, a := range items {
                    _, _ = fmt.Fprintf(tw, "%s\t%s\n", a.Code, a.City)
                }
                return tw.Flush()
            })
        },
    }
    cmd.Flags().IntVar(&limit, "limit", 50, "max items to list")
    cmd.Flags().IntVar(&offset, "offset", 0, "items to skip")
    return cmd
}

func newAirportUpdateCmd() *cobra.Command {
    var code, city string
    cmd := &cobra.Command{
        Use:   "update",
        Short: "Update an airport city by code",
        RunE: func(cmd *cobra.Command, args []string) error {
            return withAirportUsecase(func(u *usecase.AirportUsecase) error {
                if err := u.Update(context.Background(), code, city); err != nil { return err }
                fmt.Printf("updated airport %s -> %s\n", code, city)
                return nil
            })
        },
    }
    cmd.Flags().StringVar(&code, "code", "", "airport code")
    cmd.Flags().StringVar(&city, "city", "", "new city name")
    _ = cmd.MarkFlagRequired("code")
    _ = cmd.MarkFlagRequired("city")
    return cmd
}

func newAirportDeleteCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "delete <code>",
        Short: "Delete an airport by code",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            code := args[0]
            return withAirportUsecase(func(u *usecase.AirportUsecase) error {
                if err := u.Delete(context.Background(), code); err != nil { return err }
                fmt.Printf("deleted airport %s\n", code)
                return nil
            })
        },
    }
    return cmd
}
