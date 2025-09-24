package cli

import (
    "context"
    "database/sql"
    "fmt"
    "time"

    _ "github.com/jackc/pgx/v5/stdlib"
    "github.com/jmoiron/sqlx"

    "github.com/ambiyansyah-risyal/flight-booking/internal/config"
    "github.com/spf13/cobra"
)

func newDBPingCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "db:ping",
        Short: "Ping the database using current configuration",
        RunE: func(cmd *cobra.Command, args []string) error {
            cfg, err := config.Load()
            if err != nil {
                return err
            }
            dsn := cfg.Database.DSN()
            db, err := sqlx.Open("pgx", dsn)
            if err != nil {
                return fmt.Errorf("open db: %w", err)
            }
            defer func() { _ = db.Close() }()

            // Ping with timeout
            ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
            defer cancel()
            if err := db.DB.PingContext(ctx); err != nil {
                return fmt.Errorf("ping: %w", err)
            }
            // Simple sanity query
            var now time.Time
            if err := db.QueryRowContext(ctx, "select now()").Scan(&now); err != nil && err != sql.ErrNoRows {
                return fmt.Errorf("select now(): %w", err)
            }
            fmt.Printf("database reachable at %s (server time: %s)\n", cfg.Database.Host, now.Format(time.RFC3339))
            return nil
        },
    }
}

