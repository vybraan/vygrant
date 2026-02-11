package cmd

import (
	"fmt"
	"io"
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

func runClientCommandWithStdin(cmd string) {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	output, err := client.SendCommandWithPayload(cmd, input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(output)
}
