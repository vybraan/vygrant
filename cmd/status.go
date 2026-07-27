package cmd

import (
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication status for all configured accounts",
	Run: func(cmd *cobra.Command, args []string) {
		runClientCommand("status")
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
