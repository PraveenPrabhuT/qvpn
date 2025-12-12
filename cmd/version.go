package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is injected at build time via ldflags.
// Default is "dev" for local runs.
var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the current version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("VPN version %s\n", Version)
	},
}

func init() {
	// Register the command with the root
	rootCmd.AddCommand(versionCmd)
}
