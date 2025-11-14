//go:build darwin && !ios

package clip

import (
	"errors"
	"os/exec"

	"github.com/rwinkhart/go-boilerplate/back"
)

// CopyString copies a string to the clipboard.
func CopyString(clearClipboardAutomatically bool, copySubject string) error {
	cmd := exec.Command("pbcopy")
	_ = back.WriteToStdin(cmd, copySubject)
	err := cmd.Run()
	if err != nil {
		return errors.New("unable to copy to clipboard: " + err.Error())
	}
	if clearClipboardAutomatically {
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
