package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "vygrant",
	Short: "OAuth2 authentication daemon",
	Long: `A Oauth manager that handles authentication for legacy applications 
	that do not support modern standards.
	Complete documentation is available at https://github.com/vybraan/vygrant`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
