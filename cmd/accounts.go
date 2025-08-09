package cmd

import (
	"github.com/spf13/cobra"
)

var accountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "List all configured accounts",
	Long: `Displays the names of all accounts configured in vygrant.
This is useful for checking which accounts are available to run 
commands against or to authenticate.`,
	Run: func(cmd *cobra.Command, args []string) {
		runClientCommand("accounts")
	},
}

func init() {
	rootCmd.AddCommand(accountsCmd)
}
