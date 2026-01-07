package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vybraan/vygrant/internal/certgen"
)

var trustCmd = &cobra.Command{
	Use:   "trust",
	Short: "Trust vygrant local CA certificate",
	Long:  "Adds the vygrant local CA certificate to the OS or user trust store.",
	Run: func(cmd *cobra.Command, args []string) {
		system, _ := cmd.Flags().GetBool("system")
		printOnly, _ := cmd.Flags().GetBool("print")

		certPath, err := certgen.EnsureLocalCA()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to ensure local CA: %v\n", err)
			os.Exit(1)
		}

		var installErr error
		switch runtime.GOOS {
		case "darwin":
			installErr = trustDarwin(certPath, system, printOnly)
		case "windows":
			installErr = trustWindows(certPath, printOnly)
		default:
			if system {
				installErr = trustLinuxSystem(certPath, printOnly)
			} else {
				installErr = trustLinuxUser(certPath, printOnly)
			}
		}

		if installErr != nil {
			fmt.Fprintf(os.Stderr, "trust failed: %v\n", installErr)
			os.Exit(1)
		}
	},
}

func init() {
	trustCmd.Flags().Bool("system", false, "install into system trust store (may require sudo)")
	trustCmd.Flags().Bool("print", false, "print the commands without running them")
	rootCmd.AddCommand(trustCmd)
}

func trustDarwin(certPath string, system bool, printOnly bool) error {
	keychain := filepath.Join(os.Getenv("HOME"), "Library/Keychains/login.keychain-db")
	if system {
		keychain = "/Library/Keychains/System.keychain"
	}
	args := []string{"add-trusted-cert", "-d", "-r", "trustRoot", "-k", keychain, certPath}
	if printOnly {
		return printCommand("security", args...)
	}
	if system {
		return runWithSudoIfNeeded("security", args...)
	}
	return runCmd("security", args...)
}

func trustWindows(certPath string, printOnly bool) error {
	args := []string{"-user", "-addstore", "Root", certPath}
	if printOnly {
		return printCommand("certutil", args...)
	}
	return runCmd("certutil", args...)
}

func trustLinuxUser(certPath string, printOnly bool) error {
	if !hasCommand("certutil") && !printOnly {
		return fmt.Errorf("certutil not found; try --system or install nss tools")
	}

	home := os.Getenv("HOME")
	nssDir := filepath.Join(home, ".pki", "nssdb")
	dbPath := "sql:" + nssDir
	if printOnly {
		if _, err := os.Stat(filepath.Join(nssDir, "cert9.db")); os.IsNotExist(err) {
			if err := printCommand("certutil", "-d", dbPath, "-N", "--empty-password"); err != nil {
				return err
			}
		}
		return printCommand("certutil", "-d", dbPath, "-A", "-n", "vygrant local CA", "-t", "C,,", "-i", certPath)
	}

	if err := os.MkdirAll(nssDir, 0o700); err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(nssDir, "cert9.db")); os.IsNotExist(err) {
		if err := runCmd("certutil", "-d", dbPath, "-N", "--empty-password"); err != nil {
			return err
		}
	}
	return runCmd("certutil", "-d", dbPath, "-A", "-n", "vygrant local CA", "-t", "C,,", "-i", certPath)
}

func trustLinuxSystem(certPath string, printOnly bool) error {
	if hasCommand("update-ca-certificates") {
		dest := "/usr/local/share/ca-certificates/vygrant_ca.crt"
		if printOnly {
			if err := printCommand("cp", certPath, dest); err != nil {
				return err
			}
			return printCommand("update-ca-certificates")
		}
		if err := runWithSudoIfNeeded("cp", certPath, dest); err != nil {
			return err
		}
		return runWithSudoIfNeeded("update-ca-certificates")
	}
	if hasCommand("update-ca-trust") {
		dest := "/etc/pki/ca-trust/source/anchors/vygrant_ca.crt"
		if printOnly {
			if err := printCommand("cp", certPath, dest); err != nil {
				return err
			}
			return printCommand("update-ca-trust", "extract")
		}
		if err := runWithSudoIfNeeded("cp", certPath, dest); err != nil {
			return err
		}
		return runWithSudoIfNeeded("update-ca-trust", "extract")
	}
	return fmt.Errorf("no system trust tool found; install ca-certificates tools or use --print")
}

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runWithSudoIfNeeded(name string, args ...string) error {
	if runtime.GOOS == "windows" || isRoot() {
		return runCmd(name, args...)
	}
	if !hasCommand("sudo") {
		return fmt.Errorf("sudo not available; rerun with root permissions")
	}
	return runCmd("sudo", append([]string{name}, args...)...)
}

func hasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func printCommand(name string, args ...string) error {
	fmt.Printf("%s %s\n", name, strings.Join(args, " "))
	return nil
}
