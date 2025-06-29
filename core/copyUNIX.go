//go:build !windows && !darwin && !android && !ios && !termux && !wsl

package core

import (
	"errors"
	"os"
	"os/exec"

	"github.com/rwinkhart/go-boilerplate/back"
)

// copyString copies a string to the clipboard.
func copyString(continuous bool, copySubject string) error {
	// determine whether to use wl-copy (Wayland) or xclip (X11)
	var envSet, isWayland bool // track whether environment variables are set
	var cmdCopy *exec.Cmd
	if _, envSet = os.LookupEnv("WAYLAND_DISPLAY"); envSet {
		cmdCopy = exec.Command("wl-copy", "-t", "text/plain")
		isWayland = true
	} else if _, envSet = os.LookupEnv("DISPLAY"); envSet {
		cmdCopy = exec.Command("xclip", "-sel", "c", "-t", "text/plain")
	} else {
		return errors.New("clipboard platform could not be determined")
	}

	_ = back.WriteToStdin(cmdCopy, copySubject)
	err := cmdCopy.Run()
	if err != nil {
		return errors.New("unable to copy to clipboard: " + err.Error())
	}
	if !continuous {
		LaunchClipClearProcess(copySubject, isWayland)
	}
	return nil
}

// getClipCommands returns the commands for pasting and clearing the clipboard contents.
func getClipCommands() (*exec.Cmd, *exec.Cmd) {
	if os.Args[2] == "true" { // wayland
		return exec.Command("wl-paste"), exec.Command("wl-copy", "-c")
	}
	return exec.Command("xclip", "-o", "-sel", "c"), exec.Command("xclip", "-i", "/dev/null", "-sel", "c")
}
