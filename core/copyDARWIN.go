//go:build darwin

package core

import (
	"os/exec"
)

// copyString copies a string to the clipboard.
func copyString(continuous bool, copySubject string) {
	cmd := exec.Command("pbcopy")
	WriteToStdin(cmd, copySubject)
	err := cmd.Run()
	if err != nil {
		PrintError("Failed to copy to clipboard: "+err.Error(), ErrorClipboard, true)
	}

	if !continuous {
		LaunchClipClearProcess(copySubject)
	}
}

// getClipCommands returns the commands for pasting and clearing the clipboard contents.
func getClipCommands() (*exec.Cmd, *exec.Cmd) {
	cmdClear := exec.Command("pbcopy")
	WriteToStdin(cmdClear, "")
	return exec.Command("pbpaste"), cmdClear
}
