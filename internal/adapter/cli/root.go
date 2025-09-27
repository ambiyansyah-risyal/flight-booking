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
	cmd.AddCommand(newAirplaneCmd())
	cmd.AddCommand(newRouteCmd())
	cmd.AddCommand(newScheduleCmd())
	cmd.AddCommand(newBookingCmd())

	return cmd
}

// ExitHandler is a function that handles exiting the application
type ExitHandler func(int)

// DefaultExitHandler exits the application with the given code
func DefaultExitHandler(code int) {
	os.Exit(code)
}

// Execute runs the root command.
func Execute() error {
	return ExecuteWithExitHandler(DefaultExitHandler)
}

// ExecuteWithExitHandler runs the root command with a custom exit handler for testing.
func ExecuteWithExitHandler(exitHandler ExitHandler) error {
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
		exitHandler(1) // Use the exit handler instead of os.Exit directly
		return err      // Also return the error for testability
	}

	return newRootCmd().Execute()
}