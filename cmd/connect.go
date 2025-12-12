package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"

	"github.com/spf13/cobra"
)

// connectCmd represents the connect command
var connectCmd = &cobra.Command{
	Use:   "connect [target]",
	Short: "Connect to a specific VPN profile",
	Long: `Connects to the specified VPN target. 
Automatic handling of:
- Resolving Profile IDs via Regex
- Fetching Keychain passwords
- Generating 2FA/OTP tokens (unless SSO is enabled)`,
	Example: `  vpn connect dev
  vpn connect drive-prod`,

	// This enables Tab Autocompletion for the targets!
	ValidArgs: []string{"dev", "prod", "drive-dev", "drive-prod", "life"},

	// Ensure exactly one argument is passed
	Args: cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		targetName := args[0]
		performConnect(targetName)
	},
}

func init() {
	rootCmd.AddCommand(connectCmd)
}

// performConnect contains the main business logic
func performConnect(targetName string) {
	// 1. Validate Target
	config, exists := VPNTargets[targetName]
	if !exists {
		color.Red("❌ Unknown target: '%s'. Available targets: dev, prod, drive-dev, drive-prod, life\n", targetName)
		os.Exit(1)
	}

	fmt.Printf("🔍 Resolving ID for target: %s...\n", targetName)

	// 2. Resolve the Pritunl ID dynamically
	id, err := getProfileID(config.Regex)
	if err != nil {
		color.Red("❌ Failed to find profile: %v", err)
		os.Exit(1)
	}

	// 3. Prepare the command arguments
	cmdArgs := []string{"start", id}

	// 4. Handle Auth (SSO vs OTP)
	if config.SSO {
		fmt.Println("🔑 Target uses SSO. Skipping OTP generation.")
	} else {
		fmt.Printf("🔐 Fetching credentials & generating OTP for label: %s...\n", config.CotpLabel)

		otp, err := getOTP(config.CotpLabel)
		if err != nil {
			fmt.Printf("❌ Auth Error: %v\n", err)
			os.Exit(1)
		}

		cmdArgs = append(cmdArgs, "--password", otp)
		color.Green("✅ Token generated successfully.")
	}

	// 5. Execute Connection
	fmt.Printf("🚀 Connecting to %s...\n", targetName)
	output, err := runCommand("pritunl", cmdArgs...)
	if err != nil {
		fmt.Printf("❌ Pritunl Error: %s\n", err)
		os.Exit(1)
	}
	// 2. WAIT for it to actually be connected (Max wait: 30 seconds)
	// This blocks the prompt from returning until the VPN is usable.
	err = waitForState(id, true, 30*time.Second)
	if err != nil {
		fmt.Println("\n❌ Connection timed out or failed.")
		os.Exit(1)
	}

	// 6. Success Output
	color.Green("✅ Command sent successfully.")
	fmt.Println(output)
	// UPDATE STARSHIP STATE
	UpdateStateFile()
}
