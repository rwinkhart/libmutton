//go:build android && termux

package core

import (
	"os/exec"

	"github.com/rwinkhart/go-boilerplate/back"
	"github.com/rwinkhart/libmutton/global"
)

// copyString copies a string to the clipboard.
func copyString(continuous bool, copySubject string) {
	cmd := exec.Command("termux-clipboard-set")
	back.WriteToStdin(cmd, copySubject)
	err := cmd.Run()
	if err != nil {
		back.PrintError("Failed to copy to clipboard: "+err.Error(), global.ErrorClipboard, true)
	}

	if !continuous {
		LaunchClipClearProcess(copySubject)
	}
}

// getClipCommands returns the commands for pasting and clearing the clipboard contents.
func getClipCommands() (*exec.Cmd, *exec.Cmd) {
	cmdClear := exec.Command("termux-clipboard-set")
	back.WriteToStdin(cmdClear, "")
	return exec.Command("termux-clipboard-get"), cmdClear
}
