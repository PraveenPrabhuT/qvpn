package cmd

import (
	"fmt"
	"os"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "vpn",
	Short: "The Byte-Smith's VPN Orchestrator",
	Long:  `A robust CLI tool to manage Pritunl connections with integrated 2FA.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
