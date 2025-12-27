//go:build !windows && !darwin && !android && !ios && !termux && !wsl

package clip

import (
	"errors"
	"os"
	"os/exec"

	"github.com/rwinkhart/go-boilerplate/back"
)

// CopyString copies a string to the clipboard.
func CopyString(clearClipboardAutomatically bool, copySubject string) error {
	// determine whether to use wl-copy (Wayland) or xclip (X11)
	sessionIsWayland, err := isWayland()
	if err != nil {
		return err
	}
	var cmdCopy *exec.Cmd
	if sessionIsWayland {
		cmdCopy = exec.Command("wl-copy", "-t", "text/plain")
	} else {
		cmdCopy = exec.Command("xclip", "-sel", "c", "-t", "text/plain")
	}

	_ = back.WriteToStdin(cmdCopy, copySubject)
	err = cmdCopy.Run()
	if err != nil {
		return errors.New("unable to copy to clipboard: " + err.Error())
	}
	if clearClipboardAutomatically {
		LaunchClearProcess(copySubject)
	}
	return nil
}

// getClipCommands returns the commands for pasting and clearing the clipboard contents.
func getClipCommands() (*exec.Cmd, *exec.Cmd, error) {
	sessionIsWayland, err := isWayland()
	if err != nil {
		return nil, nil, err
	}
	if sessionIsWayland {
		return exec.Command("wl-paste"), exec.Command("wl-copy", "-c"), nil
	}
	return exec.Command("xclip", "-o", "-sel", "c"), exec.Command("xclip", "-i", "/dev/null", "-sel", "c"), nil
}

// isWayland returns a boolean indicating whether the current session is Wayland.
func isWayland() (bool, error) {
	var envSet bool // track whether environment variables are set
	if _, envSet = os.LookupEnv("WAYLAND_DISPLAY"); envSet {
		return true, nil
	} else if _, envSet = os.LookupEnv("DISPLAY"); envSet {
		return false, nil
	}
	return false, errors.New("unable to detect Wayland or X11 session")
}
