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

func init() {
	rootCmd.AddCommand(tokenCmd)

	tokenCmd.AddCommand(getTokenCmd)
	tokenCmd.AddCommand(deleteTokenCmd)
	tokenCmd.AddCommand(refreshTokenCmd)
}
