package daemon

import (
	"fmt"
	"net"
	"os"
	"path"
	"strings"
)

func (d *Daemon) HandleCommand(conn net.Conn, input string) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		conn.Write([]byte("Unknown command\n"))
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

		var accounts strings.Builder
		for name, _ := range d.Config.Accounts {
			accounts.WriteString(name + "\n")
		}

		writeResponse(conn, accounts.String())
	case "status":
		var status strings.Builder
		for name := range d.Config.Accounts {
			if _, err := d.TokenStore.Get(name); err != nil {
				status.WriteString(fmt.Sprintf("%s: token missing or expired \n", name))
			} else {
				status.WriteString(fmt.Sprintf("%s: token valid \n", name))
			}
		}

		writeResponse(conn, status.String())
	case "info":

		var info strings.Builder

		info.WriteString(fmt.Sprintf("cache directory: %s\n", SOCK))

		home, _ := os.UserHomeDir()
		info.WriteString(fmt.Sprintf("config file: %s\n", path.Join(home, VYGRANT_CONFIG)))
		info.WriteString(fmt.Sprintf("Server running on:\n  HTTP Port: %s\n  HTTPS Port: %s\nHTTPS public key: %s", d.Config.HTTPListen, d.Config.HTTPSListen, d.PublicKey))

		writeResponse(conn, info.String())

	case "get-token":
		if !expectArgs(conn, parts, 2, "get-token <account_name>") {
			return
		}

		accountName := parts[1]
		// token, err := storage.LoadToken(accountName)
		token, err := d.TokenStore.Get(accountName)
		if err != nil {
			writeError(conn, "Could not retrieve token for '%s': %v", accountName, err)
			return
		}
		writeResponse(conn, token.AccessToken)
	case "delete-token":
		if !expectArgs(conn, parts, 2, "delete-token <account_name>") {
			return
		}

		account := parts[1]

		// err := storage.DeleteToken(account)
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
