package cmd

import (
	"fmt"
	"os"

	"github.com/vybraan/vygrant/internal/client"
)

func runClientCommand(cmd string) {
	output, err := client.SendCommand(cmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(output)
}
