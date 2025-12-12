# 📚 Extended Setup Guide

This guide covers the one-time setup required to enable the **Zero-Touch Automation** features of QVPN (automatic OTP generation and Keychain integration).

## 1\. Install Dependencies

You need the underlying tools that QVPN orchestrates.

```bash
# 1. Install Pritunl Client (The official CLI)
brew install --cask pritunl

# 2. Install cotp (The CLI OTP generator)
brew install cotp
```

-----

## 2\. Import Your VPN Profile

Before automating it, you need the VPN profile loaded into Pritunl.

1.  **Get the URI:** Login to your organization's Pritunl web dashboard.
2.  **Download:** Click the **"Download Profile"** button (or "Copy URI").
3.  **Import:**
      * **Option A (GUI):** Open the Pritunl app and import the `.tar` or `.ovpn` file.
      * **Option B (CLI):** Run `pritunl-client import <path-to-file>`.

*Verify it worked by running `pritunl-client list`. You should see your profile there.*

-----

## 3\. Secure the "Master Password"

`cotp` encrypts your OTP secrets (QR codes) so they aren't saved as plain text. It needs a **Master Password** to unlock them.

Instead of typing this password every time you connect, we will store it in the **macOS Keychain** so QVPN can fetch it securely.

### A. Choose a Master Password

Pick a strong password (e.g., `MySecretVaultPassword123`). You will use this in the next two steps.

### B. Add to Keychain

Run this command in your terminal. Replace `YOUR_CHOSEN_PASSWORD` with the password you just picked.

```bash
# This creates a secure entry named "cotp-master-password"
security add-generic-password \
  -a "$USER" \
  -s "cotp-master-password" \
  -w "YOUR_CHOSEN_PASSWORD"
```

*QVPN is hardcoded to look for the service name `cotp-master-password`.*

-----

## 4\. Register Your OTP Secret

Now we need to tell `cotp` about your specific VPN's "Secret Key" (the QR code string).

1.  **Get the Secret:** Go back to your Pritunl web dashboard. Look for **"Two-Step Authentication"** or **"View QR Code"**. You need the text string (usually starts with `otpauth://` or just a base32 string like `JBSWY3DPEHPK3PXP`).

2.  **Add to cotp:**
    Run the import command. It will ask for a password—**enter the Master Password you chose in Step 3**.

    ```bash
    # Syntax: cotp import <VPN_ALIAS> <SECRET_KEY>
    # Example:
    cotp import dev-otp JBSWY3DPEHPK3PXP
    ```

      * **Note the Alias:** The name you use here (e.g., `dev-otp`) must match the `CotpLabel` in your QVPN config later.

3.  **Verify:**
    Run this command to test if it works. It should print a 6-digit code.

    ```bash
    # You will need to type your master password manually for this test
    cotp code dev-otp
    ```

-----

## 5\. Configure QVPN

Finally, tell QVPN how to map your commands to these profiles.

Open `cmd/config.go` (or `cmd/helpers.go` depending on your setup) and update the `VPNTargets` map:

```go
var VPNTargets = map[string]TargetConfig{
    // COMMAND ALIAS      REGEX TO MATCH PRITUNL NAME      COTP LABEL
    "dev":        {Regex: "sso_dev_vpn",          CotpLabel: "dev-otp"},
    "prod":       {Regex: "sso_prod_vpn",         CotpLabel: "prod-otp"},
}
```

  * **`"dev"`**: This is what you type (`qvpn connect dev`).
  * **`Regex`**: Part of the name seen in `pritunl-client list`.
  * **`CotpLabel`**: The name you used in **Step 4** (`cotp import dev-otp ...`).

### 🎉 Done\!

You can now run:

```bash
qvpn connect dev
```

It will:

1.  Fetch the password from Keychain.
2.  Unlock `cotp` to generate the code.
3.  Send the code to Pritunl.
4.  Wait for connection.
5.  Update your shell prompt.

### 📋 Connection Reference Matrix

Use this table to map the correct URLs to the specific aliases and labels required by the default QVPN configuration.

| Command Alias | Pritunl Profile Name (Regex) | Cotp Label (OTP Name) | Pritunl Server URL (Download Profile) |
| :--- | :--- | :--- | :--- |
| **`dev`** | `sso_ackodevvpnusers` | `devvpn` | `https://vpn-staging.acko.in/login` |
| **`prod`** | `sso_ackoprodvpnuser` | `prodvpn` | `https://vpn.acko.in/login` |
| **`drive-dev`** | `AckoDrive-Dev` | `drivedevvpn` | `https://vpn-dev.ackodrive.com/login` |
| **`drive-prod`** | `sso_ackodrive_prod` | *N/A (SSO Enabled)* | `https://vpn-prod.ackodrive.com` |
| **`life`** | `sso_ackolifevpnusers` | *N/A (SSO Enabled)* | `https://vpn.ackolife.com` |

> **Note:** For rows marked "N/A (SSO Enabled)", you do not need to set up `cotp` or Keychains. QVPN will skip the OTP step automatically.