//go:build windows || (linux && wsl)

package core

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/rwinkhart/go-boilerplate/back"
)

// copyString copies a string to the clipboard.
func copyString(continuous bool, copySubject string) {
	cmd := exec.Command("powershell.exe", "-c", fmt.Sprintf("echo '%s' | Set-Clipboard", strings.ReplaceAll(copySubject, "'", "''")))
	err := cmd.Run()
	if err != nil {
		back.PrintError("Failed to copy to clipboard: "+err.Error(), ErrorClipboard, true)
	}

	if !continuous {
		LaunchClipClearProcess(copySubject)
	}
}

// getClipCommands returns the commands for pasting and clearing the clipboard contents.
func getClipCommands() (*exec.Cmd, *exec.Cmd) {
	return exec.Command("powershell.exe", "-c", "Get-Clipboard"), exec.Command("powershell.exe", "-c", "Set-Clipboard")
}
