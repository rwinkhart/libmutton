//go:build darwin && !ios

package core

import (
	"errors"
	"os/exec"

	"github.com/rwinkhart/go-boilerplate/back"
)

// copyString copies a string to the clipboard.
func copyString(continuous bool, copySubject string) error {
	cmd := exec.Command("pbcopy")
	_ = back.WriteToStdin(cmd, copySubject)
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
	cmdClear := exec.Command("pbcopy")
	_ = back.WriteToStdin(cmdClear, "")
	return exec.Command("pbpaste"), cmdClear
}
