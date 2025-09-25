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

func newAirplaneCmd() *cobra.Command {
    cmd := &cobra.Command{Use: "airplane", Short: "Manage airplanes"}
    cmd.AddCommand(newAirplaneCreateCmd())
    cmd.AddCommand(newAirplaneListCmd())
    cmd.AddCommand(newAirplaneUpdateCmd())
    cmd.AddCommand(newAirplaneDeleteCmd())
    return cmd
}

var (
    newAirplaneDB   = func(dsn string) (*sqlx.DB, error) { return sqlxrepo.New(dsn) }
    newAirplaneRepoF = func(db *sqlx.DB) domain.AirplaneRepository { return sqlxrepo.NewAirplaneRepository(db) }
)

func withAirplaneUsecase(run func(u *usecase.AirplaneUsecase) error) error {
    cfg, err := config.Load()
    if err != nil { return err }
    db, err := newAirplaneDB(cfg.Database.DSN())
    if err != nil { return err }
    defer func(){ _ = db.Close() }()
    repo := newAirplaneRepoF(db)
    uc := NewAirplaneUsecase(repo)
    return run(uc)
}

func NewAirplaneUsecase(r domain.AirplaneRepository) *usecase.AirplaneUsecase {
    return usecase.NewAirplaneUsecase(r)
}

func newAirplaneCreateCmd() *cobra.Command {
    var code string
    var seats int
    cmd := &cobra.Command{
        Use: "create",
        Short: "Create an airplane",
        RunE: func(cmd *cobra.Command, args []string) error {
            return withAirplaneUsecase(func(u *usecase.AirplaneUsecase) error {
                a, err := u.Create(context.Background(), code, seats)
                if err != nil { return err }
                fmt.Printf("created airplane %s (%d seats)\n", a.Code, a.SeatCapacity)
                return nil
            })
        },
    }
    cmd.Flags().StringVar(&code, "code", "", "airplane code")
    cmd.Flags().IntVar(&seats, "seats", 0, "seat capacity")
    _ = cmd.MarkFlagRequired("code")
    _ = cmd.MarkFlagRequired("seats")
    return cmd
}

func newAirplaneListCmd() *cobra.Command {
    var limit, offset int
    cmd := &cobra.Command{
        Use: "list",
        Short: "List airplanes",
        RunE: func(cmd *cobra.Command, args []string) error {
            return withAirplaneUsecase(func(u *usecase.AirplaneUsecase) error {
                items, err := u.List(context.Background(), limit, offset)
                if err != nil { return err }
                tw := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
                _, _ = fmt.Fprintln(tw, "CODE\tSEATS")
                for _, a := range items {
                    _, _ = fmt.Fprintf(tw, "%s\t%d\n", a.Code, a.SeatCapacity)
                }
                return tw.Flush()
            })
        },
    }
    cmd.Flags().IntVar(&limit, "limit", 50, "max items")
    cmd.Flags().IntVar(&offset, "offset", 0, "offset")
    return cmd
}

func newAirplaneUpdateCmd() *cobra.Command {
    var code string
    var seats int
    cmd := &cobra.Command{
        Use: "update",
        Short: "Update airplane seat capacity",
        RunE: func(cmd *cobra.Command, args []string) error {
            return withAirplaneUsecase(func(u *usecase.AirplaneUsecase) error {
                if err := u.UpdateSeats(context.Background(), code, seats); err != nil { return err }
                fmt.Printf("updated airplane %s seats -> %d\n", code, seats)
                return nil
            })
        },
    }
    cmd.Flags().StringVar(&code, "code", "", "airplane code")
    cmd.Flags().IntVar(&seats, "seats", 0, "seat capacity")
    _ = cmd.MarkFlagRequired("code")
    _ = cmd.MarkFlagRequired("seats")
    return cmd
}

func newAirplaneDeleteCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "delete <code>",
        Short: "Delete an airplane by code",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            code := args[0]
            return withAirplaneUsecase(func(u *usecase.AirplaneUsecase) error {
                if err := u.Delete(context.Background(), code); err != nil { return err }
                fmt.Printf("deleted airplane %s\n", code)
                return nil
            })
        },
    }
    return cmd
}

