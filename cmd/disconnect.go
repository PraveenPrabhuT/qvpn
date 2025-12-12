package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// disconnectCmd represents the disconnect command
var disconnectCmd = &cobra.Command{
	Use:   "disconnect [target]",
	Short: "Disconnect from a VPN profile",
	Long:  `Disconnects the specified VPN profile. Use 'all' to stop every known profile.`,
	Example: `  vpn disconnect dev
  vpn disconnect all`,

	// Autocomplete includes "all" as a valid option here
	ValidArgs: []string{"dev", "prod", "drive-dev", "drive-prod", "life", "all"},

	// Ensure exactly one argument is passed
	Args: cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		targetName := args[0]

		if targetName == "all" {
			disconnectAll()
		} else {
			performDisconnect(targetName)
		}
	},
}

func init() {
	rootCmd.AddCommand(disconnectCmd)
}

// disconnectAll iterates through every configured target and stops it
func disconnectAll() {
	fmt.Println("🛑 Disconnecting EVERYTHING...")
	for name := range VPNTargets {
		// We ignore errors here because we want to try stopping everything
		// even if one fails or isn't running.
		_ = stopProfile(name, true)
	}
	// Update state file once at the very end
	UpdateStateFile()
	fmt.Println("✅ Disconnect sequence complete.")
}

// performDisconnect handles a single target
func performDisconnect(targetName string) {
	err := stopProfile(targetName, false)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}
	// Update state file after successful disconnect
	UpdateStateFile()
}

// stopProfile is a reusable helper to stop a specific profile
func stopProfile(targetName string, silent bool) error {
	config, exists := VPNTargets[targetName]
	if !exists {
		return fmt.Errorf("unknown target: %s", targetName)
	}

	if !silent {
		fmt.Printf("🔍 Resolving ID for %s...\n", targetName)
	}

	id, err := getProfileID(config.Regex)
	if err != nil {
		// If we can't find the ID, it might not be imported.
		// If we are doing "disconnect all", we just skip it.
		if silent {
			return nil
		}
		return err
	}

	if !silent {
		fmt.Printf("🛑 Stopping %s (%s)...\n", targetName, id)
	}

	_, err = runCommand("pritunl", "stop", id)
	if err != nil {
		return fmt.Errorf("failed to stop pritunl: %v", err)
	}

	// --- POLLING LOGIC START ---
	if !silent {
		fmt.Print("⏳ Waiting for disconnection...")
	}

	// Pass 'false' because we want Connected == false
	err = waitForState(id, false, 10*time.Second)
	if err != nil {
		// Even if it times out, we assume the command was sent.
		// We just warn the user.
		if !silent {
			fmt.Println("\n⚠️  Disconnect timed out (process might be stuck), but stop signal sent.")
		}
	}
	// --- POLLING LOGIC END ---

	if !silent {
		fmt.Println("\n✅ Disconnected.")
	} else {
		fmt.Printf("✅ Stopped %s\n", targetName)
	}
	return nil
}
