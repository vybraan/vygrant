package cmd

import (
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Display daemon configuration and connection details",
	Long: `Shows the daemon's cache directory, configuration file location, 
	active server ports, and the public key fingerprint used for HTTPS connections.`,
	Run: func(cmd *cobra.Command, args []string) {
		runClientCommand("info")
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
