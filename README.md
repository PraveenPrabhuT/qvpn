# 🛡️ QVPN (Quick VPN)

> **The efficient, terminal-native wrapper for Pritunl.**
> Zero-latency status checks, automated OTP handling, and instant shell integration.

**QVPN** is a Go-based CLI tool designed to make managing **Pritunl VPN** connections painless. It automates the tedious parts of connecting (finding profile IDs, entering OTPs) and exposes the connection state to your shell prompt with zero latency.

-----

## 🚀 Features

  * **⚡ Blazing Fast:** Written in Go, compiles to a static binary.
  * **🤖 OTP Automation:** Fetches your master password from macOS Keychain and generates OTPs automatically using `cotp`. No more copy-pasting from your phone.
  * **🐚 Starship Integration:** Writes state to a generic file (`~/.vpn_active`), allowing for **\<5ms** latency in your shell prompt.
  * **🧠 Context Aware:** Connects to environments by alias (`dev`, `prod`) instead of obscure Profile IDs.
  * **❄️ Nix-Native:** built with a hermetic `flake.nix` for reproducible builds anywhere.

-----

## 📦 Installation

### Option A: Homebrew (macOS)

The easiest way to install and keep updated.

```bash
brew tap PraveenPrabhuT/homebrew-tap
brew install vpn
```

### Option B: Nix (Linux & macOS)

If you use Nix Flakes, you can run it directly or install it into your profile.

```bash
# Run without installing
nix run github:PraveenPrabhuT/qvpn -- connect dev

# Install permanently
nix profile install github:PraveenPrabhuT/qvpn
```

### Option C: Go (Manual)

```bash
git clone https://github.com/PraveenPrabhuT/qvpn.git
cd qvpn
go install .
```

-----

## 🛠️ Prerequisites

1.  **Pritunl Client:** This tool wraps the standard `pritunl-client` CLI.
2.  **cotp:** Required for OTP generation (`brew install cotp`).
3.  **macOS Keychain:** (Optional) If using the auto-OTP feature, your master password must be stored in the keychain.

-----

## ⚙️ Configuration

*Currently, VPN targets are defined in the source code. (JSON config support coming in v1.0.0)*

To add your own VPNs, edit `cmd/config.go` (or `helpers.go`):

```go
var VPNTargets = map[string]TargetConfig{
    "dev":  {Regex: "sso_dev_vpn", CotpLabel: "dev-otp"},
    "prod": {Regex: "sso_prod_vpn", CotpLabel: "prod-otp"},
}
```

-----

## 🕹️ Usage

### Connect

Connects to a profile. It will wait until the connection is fully established before exiting.

```bash
# Connect to the 'dev' alias
qvpn connect dev

# Output:
# Connecting to dev...
# .......
# ✅ Connected!
```

### Disconnect

Disconnects a specific profile or all active profiles.

```bash
qvpn disconnect dev
```

### Status

Manually check status (rarely needed if using Starship).

```bash
qvpn status
```

-----

## 🎨 Starship Integration (The "Zero Latency" Setup)

QVPN writes the current active connection alias (e.g., `DEV PROD`) to `~/.vpn_active`. We use a lightweight `cat` command to read this file instantly.

Add this to your `~/.config/starship.toml`:

```toml
[custom.vpn]
# 1. Read the file. If missing, fail silently (stderr to null).
# 2. Force exit code 0 (; true) so Starship doesn't hide the module on error.
command = "cat ~/.vpn_active 2>/dev/null; true"

# The "Fast Lane" Bash shell (approx 15ms latency)
shell = ["bash", "--noprofile", "--norc"]

# Formatting
format = "[$symbol($output )]($style)"
symbol = "🔒 "
style = "fg:black bg:yellow"
```

-----

## 🏗️ Development

This project uses **Nix Flakes** to provide a reproducible dev environment.

1.  **Enter the Dev Shell:**

    ```bash
    nix develop
    ```

    This gives you `go`, `gopls`, `gotools`, and `goreleaser` automatically.

2.  **Build Binary:**

    ```bash
    go build -o vpn main.go
    ```

3.  **Release (Dry Run):**

    ```bash
    goreleaser release --snapshot --clean
    ```

-----

## 🤝 Contributing

Pull requests are welcome\! Please ensure you run `nix build` before submitting to ensure the flake remains hermetic.

**License:** MIT