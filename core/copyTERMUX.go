//go:build android && termux

package core

import (
	"fmt"
	"os"
	"os/exec"
)

// copyString copies a string to the clipboard.
func copyString(continuous bool, copySubject string) {
	cmd := exec.Command("termux-clipboard-set")
	WriteToStdin(cmd, copySubject)
	err := cmd.Run()
	if err != nil {
		fmt.Println(AnsiError+"Failed to copy to clipboard:", err.Error()+AnsiReset)
		os.Exit(ErrorClipboard)
	}

	if !continuous {
		LaunchClipClearProcess(copySubject)
	}
}

// getClipCommands returns the commands for pasting and clearing the clipboard contents.
func getClipCommands() (*exec.Cmd, *exec.Cmd) {
	cmdClear := exec.Command("termux-clipboard-set")
	WriteToStdin(cmdClear, "")
	return exec.Command("termux-clipboard-get"), cmdClear
}
