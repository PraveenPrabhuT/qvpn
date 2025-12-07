package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// --- CONFIGURATION ---
const PritunlBinPath = "/Applications/Pritunl.app/Contents/Resources/pritunl-client"

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
	"life":       {Regex: "sso_ackolifevpnusers", CotpLabel: "lifevpn"},
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

		// Safety check: verify the binary actually exists
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
