//go:build !windows

package cmd

import "os"

func isRoot() bool {
	return os.Geteuid() == 0
}
