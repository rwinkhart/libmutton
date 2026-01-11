//go:build windows || (linux && wsl)

package clip

import (
	"errors"
	"os/exec"

	"github.com/rwinkhart/go-boilerplate/back"
)

// CopyString copies a string to the clipboard.
func CopyString(clearClipboardAutomatically bool, copySubject string) error {
	cmd := exec.Command("clip.exe")
	_ = back.WriteToStdin(cmd, copySubject)
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
	clearCMD := exec.Command("clip.exe")
	_ = back.WriteToStdin(clearCMD, "")
	return exec.Command("powershell.exe", "-c", "Get-Clipboard"), clearCMD, nil
}
