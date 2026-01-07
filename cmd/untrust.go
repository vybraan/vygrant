package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

const trustCertName = "vygrant local CA"

var untrustCmd = &cobra.Command{
	Use:   "untrust",
	Short: "Remove vygrant local CA certificate trust",
	Long:  "Removes the vygrant local CA certificate from the OS or user trust store.",
	Run: func(cmd *cobra.Command, args []string) {
		system, _ := cmd.Flags().GetBool("system")
		printOnly, _ := cmd.Flags().GetBool("print")

		var err error
		switch runtime.GOOS {
		case "darwin":
			err = untrustDarwin(system, printOnly)
		case "windows":
			err = untrustWindows(printOnly)
		default:
			if system {
				err = untrustLinuxSystem(printOnly)
			} else {
				err = untrustLinuxUser(printOnly)
			}
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "untrust failed: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	untrustCmd.Flags().Bool("system", false, "remove from system trust store (may require sudo)")
	untrustCmd.Flags().Bool("print", false, "print the commands without running them")
	rootCmd.AddCommand(untrustCmd)
}

func untrustDarwin(system bool, printOnly bool) error {
	keychain := filepath.Join(os.Getenv("HOME"), "Library/Keychains/login.keychain-db")
	if system {
		keychain = "/Library/Keychains/System.keychain"
	}
	args := []string{"delete-certificate", "-c", trustCertName, keychain}
	if printOnly {
		return printCommand("security", args...)
	}
	if system {
		return runWithSudoIfNeeded("security", args...)
	}
	return runCmd("security", args...)
}

func untrustWindows(printOnly bool) error {
	args := []string{"-user", "-delstore", "Root", trustCertName}
	if printOnly {
		return printCommand("certutil", args...)
	}
	return runCmd("certutil", args...)
}

func untrustLinuxUser(printOnly bool) error {
	if !hasCommand("certutil") && !printOnly {
		return fmt.Errorf("certutil not found; try --system or install nss tools")
	}

	home := os.Getenv("HOME")
	nssDir := filepath.Join(home, ".pki", "nssdb")
	dbPath := "sql:" + nssDir
	if printOnly {
		return printCommand("certutil", "-d", dbPath, "-D", "-n", trustCertName)
	}
	return runCmd("certutil", "-d", dbPath, "-D", "-n", trustCertName)
}

func untrustLinuxSystem(printOnly bool) error {
	if hasCommand("update-ca-certificates") {
		dest := "/usr/local/share/ca-certificates/vygrant_ca.crt"
		if printOnly {
			if err := printCommand("rm", "-f", dest); err != nil {
				return err
			}
			return printCommand("update-ca-certificates")
		}
		if err := runWithSudoIfNeeded("rm", "-f", dest); err != nil {
			return err
		}
		return runWithSudoIfNeeded("update-ca-certificates")
	}
	if hasCommand("update-ca-trust") {
		dest := "/etc/pki/ca-trust/source/anchors/vygrant_ca.crt"
		if printOnly {
			if err := printCommand("rm", "-f", dest); err != nil {
				return err
			}
			return printCommand("update-ca-trust", "extract")
		}
		if err := runWithSudoIfNeeded("rm", "-f", dest); err != nil {
			return err
		}
		return runWithSudoIfNeeded("update-ca-trust", "extract")
	}
	return fmt.Errorf("no system trust tool found; install ca-certificates tools or use --print")
}
