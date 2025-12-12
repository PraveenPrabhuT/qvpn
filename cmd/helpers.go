package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// --- CONFIGURATION ---
const PritunlBinPath = "/Applications/Pritunl.app/Contents/Resources/pritunl-client"
const StateFileName = ".vpn_active"

// TargetConfig defines the rules for a specific VPN connection
type TargetConfig struct {
	Regex     string // Regex to find the profile ID
	CotpLabel string // Label to look up in 'cotp'
	SSO       bool   // If true, we skip the OTP generation step
}

// VPNTargets is the central registry of your VPNs
var VPNTargets = map[string]TargetConfig{
	"dev":        {Regex: "sso_ackodevvpnusers", CotpLabel: "devvpn"},
	"prod":       {Regex: "sso_ackoprodvpnuser", CotpLabel: "prodvpn"},
	"drive-dev":  {Regex: "AckoDrive-Dev", CotpLabel: "drivedevvpn"},
	"drive-prod": {Regex: "sso_ackodrive_prod", CotpLabel: "", SSO: true},
	"life":       {Regex: "sso_ackolifevpnusers", CotpLabel: "", SSO: true},
}

// Profile matches the JSON output structure of 'pritunl list -j'
type Profile struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Status        string `json:"status"`         // This is actually the Uptime in Pritunl's JSON
	ClientAddress string `json:"client_address"` // The VPN IP
	Connected     bool   `json:"connected"`
}

// --- HELPER FUNCTIONS ---

// getProfileID searches the Pritunl profile list for a name matching the regex
func getProfileID(pattern string) (string, error) {
	// Execute 'pritunl list -j'
	output, err := runCommand("pritunl", "list", "-j")
	if err != nil {
		return "", err
	}

	var profiles []Profile
	if err := json.Unmarshal([]byte(output), &profiles); err != nil {
		return "", fmt.Errorf("JSON parse error: %v", err)
	}

	re, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		return "", fmt.Errorf("invalid regex: %v", err)
	}

	for _, p := range profiles {
		if re.MatchString(p.Name) {
			return p.ID, nil
		}
	}
	return "", fmt.Errorf("no profile found matching '%s'", pattern)
}

// getOTP fetches the master password from Keychain and pipes it into 'cotp'
func getOTP(label string) (string, error) {
	// 1. Fetch Master Password from macOS Keychain
	// Note: Ensure the item name "cotp-master-password" matches your keychain entry exactly
	cmdKey := exec.Command("security", "find-generic-password", "-a", os.Getenv("USER"), "-s", "cotp-master-password", "-w")
	var passOut bytes.Buffer
	cmdKey.Stdout = &passOut

	if err := cmdKey.Run(); err != nil {
		return "", fmt.Errorf("failed to get password from keychain: %v", err)
	}

	password := strings.TrimSpace(passOut.String())
	if password == "" {
		return "", fmt.Errorf("keychain returned empty password")
	}

	// 2. Pipe Password into 'cotp' to generate token
	cmdCotp := exec.Command("cotp", "--password-stdin", "extract", "--label", label)
	cmdCotp.Stdin = strings.NewReader(password) // Securely pipe stdin

	var otpOut bytes.Buffer
	cmdCotp.Stdout = &otpOut

	if err := cmdCotp.Run(); err != nil {
		return "", fmt.Errorf("cotp execution failed: %v", err)
	}

	return strings.TrimSpace(otpOut.String()), nil
}

// runCommand is a wrapper to execute shell commands and capture output
func runCommand(name string, args ...string) (string, error) {
	// If the command is "pritunl", swap it for the full path
	cmdName := name
	if name == "pritunl" {
		cmdName = PritunlBinPath

		// Safety check: verify the binary actually
		if _, err := os.Stat(cmdName); os.IsNotExist(err) {
			return "", fmt.Errorf("pritunl binary not found at %s", cmdName)
		}
	}

	cmd := exec.Command(cmdName, args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%v: %s", err, stderr.String())
	}
	return strings.TrimSpace(out.String()), nil
}

// UpdateStateFile queries Pritunl and writes the active profile aliases to disk
func UpdateStateFile() {
	// 1. Get current real status from Pritunl
	// We reuse your existing logic here, but focused on finding connected ones
	output, err := runCommand("pritunl", "list", "-j")
	if err != nil {
		return // Silently fail to avoid breaking CLI flow
	}

	var profiles []Profile
	if err := json.Unmarshal([]byte(output), &profiles); err != nil {
		return
	}

	// 2. Load Config to map IDs back to Aliases (dev, prod, etc.)
	config := VPNTargets

	var activeAliases []string

	for _, p := range profiles {
		if p.Connected {
			// Try to find the alias for this ID
			foundAlias := ""
			for alias, cfg := range config {
				// We match loosely based on the Regex or Name
				// Ideally, you'd match the exact ID, but we only store Regex.
				// Re-running regex match here is safe enough.
				matched, _ := regexp.MatchString("(?i)"+cfg.Regex, p.Name)
				if matched {
					foundAlias = alias
					break
				}
			}

			// If we found a configured alias, use it. Otherwise use the raw name.
			if foundAlias != "" {
				activeAliases = append(activeAliases, strings.ToUpper(foundAlias))
			} else {
				activeAliases = append(activeAliases, "UNKNOWN")
			}
		}
	}

	// 3. Write to ~/.vpn_active
	home, _ := os.UserHomeDir()
	statePath := filepath.Join(home, StateFileName)

	if len(activeAliases) == 0 {
		// Remove file if no VPNs connected (keeps Starship clean)
		_ = os.Remove(statePath)
	} else {
		// Write "DEV PROD"
		data := strings.Join(activeAliases, " ")
		_ = os.WriteFile(statePath, []byte(data), 0644)
	}
}

// waitForState polls Pritunl until the profile with 'profileID' matches the 'desiredConnected' state.
// It times out after 'timeout' duration.
func waitForState(profileID string, desiredConnected bool, timeout time.Duration) error {
	ticker := time.NewTicker(500 * time.Millisecond) // Check every 0.5s
	defer ticker.Stop()

	timeoutChan := time.After(timeout)

	for {
		select {
		case <-timeoutChan:
			return fmt.Errorf("timeout waiting for VPN state change")
		case <-ticker.C:
			// Fetch status
			output, err := runCommand("pritunl", "list", "-j")
			if err != nil {
				continue // Ignore transient errors during polling
			}

			var profiles []Profile
			if err := json.Unmarshal([]byte(output), &profiles); err != nil {
				continue
			}

			// Find our profile
			for _, p := range profiles {
				if p.ID == profileID {
					if p.Connected == desiredConnected {
						return nil // Success! State matched.
					}
					// Optional: Print a dot to show aliveness
					fmt.Print(".")
				}
			}
		}
	}
}
