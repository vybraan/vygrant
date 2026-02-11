package cmd

import (
	"github.com/spf13/cobra"
)

var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Manage OAuth2 tokens",
	Long:  `Allows listing, retrieving, and deleting stored OAuth2 tokens.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var getTokenCmd = &cobra.Command{
	Use:   "get [account_name]",
	Short: "Get a specific token",
	Long:  `Retrieves and displays the token for a specified account.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		accountName := args[0]
		runClientCommand("get-token " + accountName)
	},
}

var deleteTokenCmd = &cobra.Command{
	Use:   "delete [account_name]",
	Short: "Delete a specific token",
	Long:  `Deletes the token associated with a specified account.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		accountName := args[0]
		runClientCommand("delete-token " + accountName)
	},
}

var refreshTokenCmd = &cobra.Command{
	Use:   "refresh [account_name]",
	Short: "Refresh a specific token",
	Long:  `Refreshes the token for the specified account.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		accountName := args[0]
		runClientCommand("refresh-token " + accountName)
	},
}

var dumpTokenCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump token state to stdout (sensitive)",
	Long:  "Dumps the current token state to stdout. Treat this output as sensitive and encrypt it.",
	Run: func(cmd *cobra.Command, args []string) {
		runClientCommand("dump-tokens")
	},
}

var restoreTokenCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore token state from stdin (sensitive)",
	Long:  "Restores token state from stdin. Treat the input as sensitive.",
	Run: func(cmd *cobra.Command, args []string) {
		runClientCommandWithStdin("restore-tokens")
	},
}

func init() {
	rootCmd.AddCommand(tokenCmd)

	tokenCmd.AddCommand(getTokenCmd)
	tokenCmd.AddCommand(deleteTokenCmd)
	tokenCmd.AddCommand(refreshTokenCmd)
	tokenCmd.AddCommand(dumpTokenCmd)
	tokenCmd.AddCommand(restoreTokenCmd)
}
