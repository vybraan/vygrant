package client

import (
	"bytes"
	"fmt"
	"net"
	"strings"

	"github.com/vybraan/vygrant/internal/daemon"
)

// SendCommand sends a text command to the daemon over the Unix domain socket.
// It returns the daemon's response with leading and trailing whitespace removed,
// or an error if connecting, writing, or reading from the socket fails.
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

func SendCommandWithPayload(command string, payload []byte) (string, error) {
	conn, err := net.Dial("unix", daemon.SocketPath())
	if err != nil {
		return "", fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer conn.Close()

	var buf bytes.Buffer
	buf.WriteString(command)
	buf.WriteByte('\n')
	if len(payload) > 0 {
		buf.Write(payload)
	}

	if _, err := conn.Write(buf.Bytes()); err != nil {
		return "", fmt.Errorf("failed to send command: %w", err)
	}

	resp := make([]byte, 8192)
	n, err := conn.Read(resp)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}
	return strings.TrimSpace(string(resp[:n])), nil
}
