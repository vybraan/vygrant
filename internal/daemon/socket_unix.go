//go:build !windows

package daemon

import (
	"os"
	"path/filepath"
	"strconv"
)

const socketName = "vygrant.sock"

// SocketPath returns the filesystem path to the daemon's Unix socket file.
// It prefers $XDG_RUNTIME_DIR/vygrant.sock when XDG_RUNTIME_DIR is set and is a directory,
// falls back to /run/user/<uid>/vygrant.sock when that directory exists, and otherwise
// uses the system temporary directory (os.TempDir()) with vygrant.sock.
func SocketPath() string {
	if runtimeDir := os.Getenv("XDG_RUNTIME_DIR"); runtimeDir != "" {
		if info, err := os.Stat(runtimeDir); err == nil && info.IsDir() {
			return filepath.Join(runtimeDir, socketName)
		}
	}

	runtimeDir := filepath.Join("/run/user", strconv.Itoa(os.Getuid()))
	if _, err := os.Stat(runtimeDir); err == nil {
		return filepath.Join(runtimeDir, socketName)
	}

	return filepath.Join(os.TempDir(), socketName)
}