//go:build android && termux

package clip

import (
	"errors"
	"os/exec"

	"github.com/rwinkhart/go-boilerplate/back"
)

// CopyString copies a string to the clipboard.
func CopyString(continuous bool, copySubject string) error {
	cmd := exec.Command("termux-clipboard-set")
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
	cmdClear := exec.Command("termux-clipboard-set")
	_ = back.WriteToStdin(cmdClear, "")
	return exec.Command("termux-clipboard-get"), cmdClear
}
