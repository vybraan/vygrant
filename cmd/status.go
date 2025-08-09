package cmd

import (
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication status for all configured accounts",
	Long: `Checks the authentication tokens for all configured accounts 
	and reports whether they are valid, expired, or missing.`,
	Run: func(cmd *cobra.Command, args []string) {
		runClientCommand("status")
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
