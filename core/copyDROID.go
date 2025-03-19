//go:build android && !termux

package core

import (
	"os/exec"
)

// copyString copies a string to the clipboard.
func copyString(continuous bool, copySubject string) {
	// temporary dummy function to allow Android compilation
}

// getClipCommands returns the commands for pasting and clearing the clipboard contents.
func getClipCommands() (*exec.Cmd, *exec.Cmd) {
	// temporary dummy function to allow Android compilation
	return nil, nil
}
