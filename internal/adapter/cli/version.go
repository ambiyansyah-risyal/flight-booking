package cli

import (
    "fmt"

    "github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "version",
        Short: "Print version information",
        Run: func(cmd *cobra.Command, args []string) {
            if BuildDate != "" {
                fmt.Printf("flight-booking %s (%s) %s\n", Version, Commit, BuildDate)
            } else {
                fmt.Printf("flight-booking %s (%s)\n", Version, Commit)
            }
        },
    }
}
