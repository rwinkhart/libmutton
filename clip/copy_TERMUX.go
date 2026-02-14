//go:build android && termux

package clip

import (
	"errors"
	"os/exec"

	"github.com/rwinkhart/go-boilerplate/back"
)

// CopyBytes copies a byte slice to the clipboard.
func CopyBytes(clearClipboardAutomatically bool, copySubject []byte) error {
	cmd := exec.Command("termux-clipboard-set")
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
	cmdClear := exec.Command("termux-clipboard-set")
	_ = back.WriteToStdinAndZeroizeInput(cmdClear, nil)
	return exec.Command("termux-clipboard-get"), cmdClear, nil
}
