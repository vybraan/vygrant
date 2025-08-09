package daemon

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
)

func Notify(title, message string) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("notify-send", title, message)
	case "darwin":
		script := fmt.Sprintf(`display notification %q with title %q`, message, title)
		cmd = exec.Command("osascript", "-e", script)
	case "windows":
		ps := fmt.Sprintf(`[reflection.assembly]::LoadWithPartialName('System.Windows.Forms');`+
			`[System.Windows.Forms.MessageBox]::Show('%s','%s')`, message, title)
		cmd = exec.Command("powershell", "-Command", ps)
	}

	if cmd != nil {
		if err := cmd.Run(); err != nil {
			log.Printf("notification failed: %v", err)
		}
	}
}
