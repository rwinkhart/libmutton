//go:build android && termux

package core

import (
	"os/exec"
)

// copyString copies a string to the clipboard.
func copyString(continuous bool, copySubject string) {
	cmd := exec.Command("termux-clipboard-set")
	writeToStdin(cmd, copySubject)
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
	cmdClear := exec.Command("termux-clipboard-set")
	writeToStdin(cmdClear, "")
	return exec.Command("termux-clipboard-get"), cmdClear
}
