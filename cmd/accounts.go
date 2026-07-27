package cmd

import (
	"github.com/spf13/cobra"
)

var accountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "List all configured accounts",
	Run: func(cmd *cobra.Command, args []string) {
		runClientCommand("accounts")
	},
}

func init() {
	rootCmd.AddCommand(accountsCmd)
}
