package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Vygrant",
	Long:  `All software has versions. This is Vygrant's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Vygrant - OAuth2 authentication daemon %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
