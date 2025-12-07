package cmd

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

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
		fmt.Printf("❌ Error fetching status: %v\n", err)
		return
	}

	// 2. Parse JSON
	var profiles []Profile
	if err := json.Unmarshal([]byte(output), &profiles); err != nil {
		fmt.Printf("❌ JSON Parse Error: %v\n", err)
		return
	}

	if len(profiles) == 0 {
		fmt.Println("🤷 No Pritunl profiles found.")
		return
	}

	// 3. Render Dashboard
	fmt.Println("--- 🌐 VPN Connection Status ---")

	for _, p := range profiles {
		printProfile(p)
	}

	fmt.Println("--------------------------------")
}

func printProfile(p Profile) {
	// --- Name Cleanup Logic ---
	// Replicating your JQ logic: sub(".*sso_"; "") | sub("vpnusers.*"; "") | ascii_upcase
	cleanName := p.Name

	// Remove prefix ".*sso_"
	rePrefix := regexp.MustCompile(`(?i).*sso_`)
	cleanName = rePrefix.ReplaceAllString(cleanName, "")

	// Remove suffix "vpnusers.*"
	reSuffix := regexp.MustCompile(`(?i)vpnusers.*`)
	cleanName = reSuffix.ReplaceAllString(cleanName, "")

	// Uppercase
	cleanName = strings.ToUpper(cleanName)

	// --- Visuals ---
	icon := "❌"
	color := "\033[31m" // Red
	statusText := "DISCONNECTED"
	reset := "\033[0m"

	if p.Connected {
		icon = "✅"
		color = "\033[32m" // Green
		statusText = "CONNECTED"
	}

	// Print Header (Icon + Name)
	fmt.Printf("%s %s▶️  Profile: %s%s\n", color, icon, cleanName, reset)

	// Print Details
	fmt.Printf("   Status: %s%s%s\n", color, statusText, reset)

	if p.Connected {
		fmt.Printf("   ⏱️  Uptime: %s\n", p.Status)
		fmt.Printf("   💻 IP: %s\n", p.ClientAddress)
	}
	// Add a little spacer between items
	fmt.Println("")
}
