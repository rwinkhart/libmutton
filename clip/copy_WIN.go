//go:build windows || (linux && wsl)

package clip

import (
	"errors"
	"os/exec"
	"syscall"

	"github.com/rwinkhart/go-boilerplate/back"
	"github.com/rwinkhart/go-boilerplate/security"
	"golang.org/x/sys/windows"
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
	sysProcAttr := &syscall.SysProcAttr{CreationFlags: windows.CREATE_NO_WINDOW}
	pasteCMD := exec.Command("powershell.exe", "-c", "Get-Clipboard")
	pasteCMD.SysProcAttr = sysProcAttr
	clearCMD := exec.Command("clip.exe")
	clearCMD.SysProcAttr = sysProcAttr
	_ = back.WriteToStdin(clearCMD, nil, false)
	return pasteCMD, clearCMD, nil
}
