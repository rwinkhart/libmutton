//go:build windows || (linux && wsl)

package clip

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// CopyString copies a string to the clipboard.
func CopyString(continuous bool, copySubject string) error {
	cmd := exec.Command("powershell.exe", "-c", fmt.Sprintf("echo '%s' | Set-Clipboard", strings.ReplaceAll(copySubject, "'", "''")))
	err := cmd.Run()
	if err != nil {
		return errors.New("unable to copy to clipboard: " + err.Error())
	}
	if !continuous {
		LaunchClipClearProcess(copySubject)
	}
	return nil
}

// getClipCommands returns the commands for pasting and clearing the clipboard contents.
func getClipCommands() (*exec.Cmd, *exec.Cmd) {
	return exec.Command("powershell.exe", "-c", "Get-Clipboard"), exec.Command("powershell.exe", "-c", "Set-Clipboard")
}
