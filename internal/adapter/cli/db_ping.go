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

// DBConnector interface for database operations to enable testing
type DBConnector interface {
    PingContext(ctx context.Context) error
    QueryRowContext(ctx context.Context, query string) *sql.Row
    Close() error
}

// RealDBConnector wraps sqlx.DB to implement DBConnector interface
type RealDBConnector struct {
    db *sqlx.DB
}

func (r *RealDBConnector) PingContext(ctx context.Context) error {
    return r.db.PingContext(ctx)
}

func (r *RealDBConnector) QueryRowContext(ctx context.Context, query string) *sql.Row {
    return r.db.QueryRowContext(ctx, query)
}

func (r *RealDBConnector) Close() error {
    return r.db.Close()
}

// DBConnectionFactory creates database connections
type DBConnectionFactory func(driverName, dataSourceName string) (DBConnector, error)

// DefaultDBConnectionFactory creates real database connections
func DefaultDBConnectionFactory(driverName, dataSourceName string) (DBConnector, error) {
    db, err := sqlx.Open(driverName, dataSourceName)
    if err != nil {
        return nil, err
    }
    return &RealDBConnector{db: db}, nil
}

func newDBPingCmd() *cobra.Command {
    return newDBPingCmdWithFactory(DefaultDBConnectionFactory)
}

func newDBPingCmdWithFactory(factory DBConnectionFactory) *cobra.Command {
    return &cobra.Command{
        Use:   "db:ping",
        Short: "Ping the database using current configuration",
        RunE: func(cmd *cobra.Command, args []string) error {
            cfg, err := config.Load()
            if err != nil {
                return err
            }
            dsn := cfg.Database.DSN()
            
            db, err := factory("pgx", dsn)
            if err != nil {
                return fmt.Errorf("open db: %w", err)
            }
            defer func() { _ = db.Close() }()

            // Ping with timeout
            ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
            defer cancel()
            if err := db.PingContext(ctx); err != nil {
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
