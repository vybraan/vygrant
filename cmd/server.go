package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/vybraan/vygrant/internal/daemon"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run the daemon auth",
	Long:  "run the daemon to handle the oauth stuff",
	Run: func(cmd *cobra.Command, args []string) {
		daemon, err := daemon.NewDaemon()
		if err != nil {
			log.Fatalf("ERROR: %v", err)
		}
		daemon.Start()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
