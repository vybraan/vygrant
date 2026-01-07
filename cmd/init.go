package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const defaultConfigContent = `# vygrant configuration file

https_listen = "8080"
http_listen = "none"
persist_tokens = true
[account]
# [account.example]
# auth_uri = "https://example.com/oauth2/authorize"
# token_uri = "https://example.com/oauth2/token"
# client_id = "your_client_id"
# client_secret = "your_client_secret"
# redirect_uri = "https://localhost:8080"
# redirect_uri = "http://localhost:8080" # use with http_listen to avoid self-signed TLS warnings
# scopes = [
#   "profile",
#   "name",
#   "email",
#   "offline_access"
# ]
# 
# [account.example.auth_uri_fields]
# login_hint = "example@example.com"
`

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize default configuration",
	Long:  `Creates a default configuration file in your home directory (~/.config/vybr/vygrant.toml).`,
	Run: func(cmd *cobra.Command, args []string) {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to get home directory: %v\n", err)
			os.Exit(1)
		}

		configDir := filepath.Join(home, ".config", "vybr")
		configFile := filepath.Join(configDir, "vygrant.toml")

		if _, err := os.Stat(configFile); err == nil {
			fmt.Fprintf(os.Stderr, "Config file already exists at %s\n", configFile)
			os.Exit(1)
		}

		if err := os.MkdirAll(configDir, 0o700); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create config directory: %v\n", err)
			os.Exit(1)
		}

		err = os.WriteFile(configFile, []byte(defaultConfigContent), 0o600)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write config file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Default config file created at %s\n", configFile)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
