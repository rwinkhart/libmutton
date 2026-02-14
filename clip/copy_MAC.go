//go:build darwin && !ios

package clip

import (
	"errors"
	"os/exec"

	"github.com/rwinkhart/go-boilerplate/back"
)

// CopyBytes copies a byte slice to the clipboard.
func CopyBytes(clearClipboardAutomatically bool, copySubject []byte) error {
	cmd := exec.Command("pbcopy")
	_ = back.WriteToStdinAndZeroizeInput(cmd, copySubject)
	if err := cmd.Run(); err != nil {
		return errors.New("unable to copy to clipboard: " + err.Error())
	}
	if clearClipboardAutomatically {
		LaunchClearProcess(copySubject)
	}
	return nil
}

// getClipCommands returns the commands for pasting and clearing the clipboard contents.
func getClipCommands() (*exec.Cmd, *exec.Cmd, error) {
	cmdClear := exec.Command("pbcopy")
	_ = back.WriteToStdinAndZeroizeInput(cmdClear, nil)
	return exec.Command("pbpaste"), cmdClear, nil
}
