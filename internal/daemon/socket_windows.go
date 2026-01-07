//go:build windows

package daemon

import (
	"os"
	"path/filepath"
)

const socketName = "vygrant.sock"

func SocketPath() string {
	return filepath.Join(os.TempDir(), socketName)
}
