//go:build windows || (linux && wsl)

package clip

import (
	"errors"
	"os/exec"

	"github.com/rwinkhart/go-boilerplate/back"
	"github.com/rwinkhart/go-boilerplate/security"
)

// CopyBytes copies a byte slice to the clipboard.
func CopyBytes(clearClipboardAutomatically bool, copySubject []byte) error {
	cmd := exec.Command("clip.exe")
	_ = back.WriteToStdin(cmd, copySubject, false)
	if err := cmd.Run(); err != nil {
		security.ZeroizeBytes(copySubject)
		return errors.New("unable to copy to clipboard: " + err.Error())
	}
	if clearClipboardAutomatically {
		LaunchClearProcess(copySubject)
	} else {
		security.ZeroizeBytes(copySubject)
	}
	return nil
}

// getClipCommands returns the commands for pasting and clearing the clipboard contents.
func getClipCommands() (*exec.Cmd, *exec.Cmd, error) {
	clearCMD := exec.Command("clip.exe")
	_ = back.WriteToStdin(clearCMD, nil, false)
	return exec.Command("powershell.exe", "-c", "Get-Clipboard"), clearCMD, nil
}
