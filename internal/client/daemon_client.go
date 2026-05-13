package client

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/vybraan/vygrant/internal/daemon"
)

func readResponse(conn net.Conn) (string, error) {
	data, err := io.ReadAll(conn)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

func shutdownWrite(conn net.Conn) {
	if c, ok := conn.(interface{ CloseWrite() error }); ok {
		c.CloseWrite()
	}
}

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

	shutdownWrite(conn)

	return readResponse(conn)
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

	shutdownWrite(conn)

	return readResponse(conn)
}
