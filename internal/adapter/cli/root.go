package cli

import (
    "fmt"
    "os"

    "github.com/ambiyansyah-risyal/flight-booking/internal/config"
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

var (
    Version   = "0.2.0-dev"
    Commit    = "dev"
    BuildDate = ""
)

func newRootCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "flight-booking",
        Short: "CLI for managing flights, routes, schedules, and bookings",
        Run: func(cmd *cobra.Command, args []string) {
            _ = cmd.Help()
        },
        SilenceUsage:  true,
        SilenceErrors: true,
    }

    // Global flags (optional config file)
    cmd.PersistentFlags().String("config", "", "path to config file (yaml, toml, json)")
    _ = viper.BindPFlag("config", cmd.PersistentFlags().Lookup("config"))

    // Subcommands
    cmd.AddCommand(newVersionCmd())
    cmd.AddCommand(newDBPingCmd())
    cmd.AddCommand(newAirportCmd())

    return cmd
}

// Execute runs the root command.
func Execute() error {
    // Viper base setup
    viper.SetEnvPrefix("FLIGHT")
    viper.AutomaticEnv()

    // Optionally load a config file if provided
    cfgPath := viper.GetString("config")
    if cfgPath != "" {
        viper.SetConfigFile(cfgPath)
        if err := viper.ReadInConfig(); err != nil {
            return fmt.Errorf("read config: %w", err)
        }
    }

    // Initialize app config to validate env
    if _, err := config.Load(); err != nil {
        fmt.Fprintf(os.Stderr, "config error: %v\n", err)
        os.Exit(1)
    }

    return newRootCmd().Execute()
}
