package daemon

import (
	"fmt"
	"net"
	"os"
	"path"
	"strings"
	"time"
)

func (d *Daemon) HandleCommand(conn net.Conn, input string) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		writeError(conn, "No command provided")
		return
	}

	cmd := parts[0]

	switch cmd {
	case "accounts":
		if len(d.Config.Accounts) == 0 {
			writeResponse(conn, "No accounts configured.")
			return

		}

		var accountList strings.Builder
		for name := range d.Config.Accounts {
			accountList.WriteString(name + "\n")
		}
		writeResponse(conn, accountList.String())

	case "status":
		var status []string
		for name := range d.Config.Accounts {
			if _, err := d.TokenStore.Get(name); err != nil {
				status = append(status, fmt.Sprintf("%s: token missing or expired", name))
			} else {
				status = append(status, fmt.Sprintf("%s: token valid", name))
			}
		}
		writeResponse(conn, strings.Join(status, "\n"))

	case "info":

		home, _ := os.UserHomeDir()
		info := fmt.Sprintf(
			"Cache directory: %s\nConfig file: %s\nServer running on:\n  HTTP Port: %s\n  HTTPS Port: %s\nHTTPS public key: %s",
			SOCK,
			path.Join(home, VYGRANT_CONFIG),
			d.Config.HTTPListen,
			d.Config.HTTPSListen,
			d.PublicKey,
		)
		writeResponse(conn, info)

	case "get-token":
		if !expectArgs(conn, parts, 2, "get-token <account_name>") {
			return
		}

		account := parts[1]
		token, err := d.TokenStore.Get(account)

		if err != nil {
			authLink := fmt.Sprintf("https://localhost:%s/auth?account=%s", d.Config.HTTPSListen, account)
			writeError(conn, "Could not retrieve token for '%s': %v. Please authenticate. Go to: %s", account, err, authLink)
			return
		}

		// Auto-refresh token if expired
		if err == nil && token.Expiry.Before(time.Now()) && token.RefreshToken != "" {
			newToken, err := RefreshToken(account, d.Config, token)
			if err != nil {

				Notify("vygrant - auto refresh failed", fmt.Sprintf("Token for '%s' could not be refreshed and has been deleted. Please re-authenticate.", account))
				writeError(conn, "Failed to auto refresh token for '%s': %v", account, err)
				d.TokenStore.Delete(account)
				return
			}
			d.TokenStore.Set(account, newToken)
			token = newToken
			Notify("vygrant - token refreshed", fmt.Sprintf("Token for '%s' successfully refreshed.", account))
		}

		writeResponse(conn, token.AccessToken)

	case "delete-token":
		if !expectArgs(conn, parts, 2, "delete-token <account_name>") {
			return
		}

		account := parts[1]
		err := d.TokenStore.Delete(account)
		if err != nil {
			writeError(conn, "Could not delete token for '%s': %v", account, err)
			return
		}
		writeResponse(conn, "Token for '%s' deleted", account)

	case "refresh-token":
		if !expectArgs(conn, parts, 2, "refresh-token <account_name>") {
			return
		}

		account := parts[1]
		token, err := d.TokenStore.Get(account)

		if err != nil || token.RefreshToken == "" {
			authLink := fmt.Sprintf("https://localhost:%s/auth?account=%s", d.Config.HTTPSListen, account)
			Notify("vygrant - no refresh token", fmt.Sprintf("No refresh token for '%s'. Authenticate at: %s", account, authLink))
			writeError(conn, "No refresh token available for '%s'. Please authenticate at: %s", account, authLink)
			return
		}

		newToken, err := RefreshToken(account, d.Config, token)
		if err != nil {
			writeError(conn, "Failed to refresh token for '%s': %v", account, err)
			d.TokenStore.Delete(account)
			return
		}
		d.TokenStore.Set(account, newToken)
		Notify("vygrant - token refreshed", fmt.Sprintf("Token for '%s' successfully refreshed.", account))
		writeResponse(conn, "Token for '%s' refreshed", account)

	default:
		writeError(conn, "Unknown command '%s'", parts[0])
	}
}

func expectArgs(conn net.Conn, parts []string, expected int, usage string) bool {
	if len(parts) != expected {
		writeError(conn, "Invalid arguments. Usage: %s", usage)
		return false
	}
	return true
}
