//go:build windows

package daemon

import (
	"os"
	"path/filepath"
)

const socketName = "vygrant.sock"

// SocketPath returns the full path to the daemon socket file (vygrant.sock) located in the system temporary directory.
func SocketPath() string {
	return filepath.Join(os.TempDir(), socketName)
}