//go:build !windows && !darwin && !android && !termux && !wsl

package core

import (
	"os"
	"os/exec"
)

// copyString copies a string to the clipboard.
func copyString(continuous bool, copySubject string) {
	var envSet, isWayland bool // track whether environment variables are set
	var cmdCopy *exec.Cmd
	// determine whether to use wl-copy (Wayland) or xclip (X11)
	if _, envSet = os.LookupEnv("WAYLAND_DISPLAY"); envSet {
		cmdCopy = exec.Command("wl-copy", "-t", "text/plain")
		isWayland = true
	} else if _, envSet = os.LookupEnv("DISPLAY"); envSet {
		cmdCopy = exec.Command("xclip", "-sel", "c", "-t", "text/plain")
	} else {
		PrintError("Clipboard platform could not be determined", ErrorClipboard, true)
	}

	writeToStdin(cmdCopy, copySubject)
	err := cmdCopy.Run()
	if err != nil {
		PrintError("Failed to copy to clipboard: "+err.Error(), ErrorClipboard, true)
	}

	if !continuous {
		LaunchClipClearProcess(copySubject, isWayland)
	}
}

// getClipCommands returns the commands for pasting and clearing the clipboard contents.
func getClipCommands() (*exec.Cmd, *exec.Cmd) {
	if os.Args[2] == "true" { // wayland
		return exec.Command("wl-paste"), exec.Command("wl-copy", "-c")
	}
	return exec.Command("xclip", "-o", "-sel", "c"), exec.Command("xclip", "-i", "/dev/null", "-sel", "c")
}
