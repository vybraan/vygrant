package cmd

import (
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Display daemon configuration and connection details",
	Run: func(cmd *cobra.Command, args []string) {
		runClientCommand("info")
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
