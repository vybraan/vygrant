package client

import (
	"fmt"
	"net"
	"strings"

	"github.com/vybraan/vygrant/internal/daemon"
)

func SendCommand(command string) (string, error) {

	conn, err := net.Dial("unix", daemon.SocketPath())
	if err != nil {
		return "", fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte(command + "\n"))
	if err != nil {
		return "", fmt.Errorf("failed to send command: %w", err)
	}

	buf := make([]byte, 8192)
	n, err := conn.Read(buf)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return strings.TrimSpace(string(buf[:n])), nil
}
