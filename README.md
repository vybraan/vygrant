# vygrant

**OAuth2 Authentication Daemon for Legacy Applications**

`vygrant` provides a local CLI and background daemon for managing OAuth2 tokens on legacy applications that lack modern authentication support.

## Features

- **Daemon + CLI**: Manage and refresh tokens through a local Unix socket.
- **Secure token handling**: Tokens are stored securely in memory and optionally persisted.
- **Auto-refresh & notifications**: Optional background token refresh with notification support on Linux/macOS/Windows.

## Installation

### Arch Linux

You can install `vygrant` from the AUR using an AUR helper like `yay` or `paru`:

```bash
yay -S vygrant
# or
paru -S vygrant
```

Else you can install it manually using `git` and `makepkg`:

```bash
git clone https://aur.archlinux.org/vygrant.git
cd vygrant
makepkg -si
```

### Manual

```bash
git clone https://github.com/vybraan/vygrant.git
cd vygrant
go build  -ldflags "-s -w"
```

## Getting Started

### 1. Initialize Configuration

Create a default configuration file:

```bash
vygrant init
```

This generates a config at `~/.config/vybr/vygrant.toml`. Open and edit the file to register your OAuth2 accounts:

```toml
https_listen = "8443"
http_listen = ""
persist_tokens = true

[account.myapp]
auth_uri = "https://provider.com/auth"
token_uri = "https://provider.com/token"
client_id = "YOUR_CLIENT_ID"
client_secret = "YOUR_CLIENT_SECRET"
redirect_uri = "https://localhost:8443"
scopes = ["openid", "profile", "email"]
```

#### Optional Settings

- `https_listen`: Port for HTTPS callbacks (default `8080`).
- `http_listen`: Port for HTTP callbacks (default disabled).
- `persist_tokens`: Whether to save tokens to disk (default `true`).

###### You may use Thunderbird's OAuth2 client ID/secret for Microsoft accounts, but it's recommended to create your own credentials.

### 2. Start the Daemon

Ensure the config exists, then run:

```bash
vygrant server
```

The daemon will listen for OAuth2 callbacks and manage the tokens.

### 3. Authenticate via CLI

Use the CLI to initiate authentication in your browser:

```bash
vygrant token refresh myapp
```

After approval in the browser, you'll see a friendly success page. You can then close the tab and vygrant handles everything in the background.

## CLI Commands Overview

- `vygrant accounts` - list all configured accounts.
- `vygrant status` - display authentication status (valid, expired, missing).
- `vygrant info` - show daemon config details (socket path, ports, etc.).
- `vygrant token get <account>` - retrieve access token.
- `vygrant token delete <account>` - remove a stored token.
- `vygrant token refresh <account>` - perform OAuth authentication flow (opens browser).

## Example usage with msmtp

```
account example@hotmail.com
host smtp-mail.outlook.com
port 587
from example@hotmail.com
user example@hotmail.com
passwordeval "vygrant token get myapp"
auth on
tls on
tls_trust_file	/etc/ssl/certs/ca-certificates.crt
tls_starttls
```

## Contributing & License

Contributions are welcome! Please fork, submit pull requests, or file issues for enhancements or bug fixes.

_vygrant_ is released under the MIT License.
