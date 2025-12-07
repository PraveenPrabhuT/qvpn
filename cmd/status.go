package cmd

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show VPN connection status",
	Long:  `Displays a dashboard of all configured VPN profiles, their connection state, uptime, and client IP address.`,
	Run: func(cmd *cobra.Command, args []string) {
		showStatus()
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func showStatus() {
	// 1. Fetch raw JSON from Pritunl
	output, err := runCommand("pritunl", "list", "-j")
	if err != nil {
		// Use color.Red for errors
		color.Red("❌ Error fetching status: %v", err)
		return
	}

	// 2. Parse JSON
	var profiles []Profile
	if err := json.Unmarshal([]byte(output), &profiles); err != nil {
		color.Red("❌ JSON Parse Error: %v", err)
		return
	}

	if len(profiles) == 0 {
		fmt.Println("🤷 No Pritunl profiles found.")
		return
	}

	// 3. Render Dashboard
	// Print a bold header
	color.New(color.Bold).Println("--- 🌐 VPN Connection Status ---")

	for _, p := range profiles {
		printProfile(p)
	}

	color.New(color.Bold).Println("--------------------------------")
}

func printProfile(p Profile) {
	// --- Name Cleanup Logic (Same as before) ---
	cleanName := p.Name
	rePrefix := regexp.MustCompile(`(?i).*sso_`)
	cleanName = rePrefix.ReplaceAllString(cleanName, "")
	reSuffix := regexp.MustCompile(`(?i)vpnusers.*`)
	cleanName = reSuffix.ReplaceAllString(cleanName, "")
	cleanName = strings.ToUpper(cleanName)

	// --- Visuals with fatih/color ---

	// Create reusable color printers
	// .SprintFunc() returns a function that wraps text in that color
	green := color.New(color.FgGreen, color.Bold).SprintFunc()
	red := color.New(color.FgRed, color.Bold).SprintFunc()
	faint := color.New(color.Faint).SprintFunc() // Grey-ish for labels

	if p.Connected {
		// ✅ CONNECTED STATE
		// Format: Icon + Green Arrow + Green Name
		fmt.Printf("✅ %s Profile: %s\n", green("▶️"), green(cleanName))

		fmt.Printf("   Status: %s\n", green("CONNECTED"))
		fmt.Printf("   %s %s\n", faint("⏱️  Uptime:"), p.Status)
		fmt.Printf("   %s %s\n", faint("💻 IP:"), p.ClientAddress)
	} else {
		// ❌ DISCONNECTED STATE
		// Format: Icon + Red Arrow + Red Name
		fmt.Printf("❌ %s Profile: %s\n", red("▶️"), red(cleanName))

		fmt.Printf("   Status: %s\n", red("DISCONNECTED"))
	}

	// Spacer line
	fmt.Println("")
}
