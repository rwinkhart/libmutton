//go:build !windows && !darwin && !android && !termux && !wsl && !interactive

package core

import (
	"os"
	"os/exec"
	"strconv"
)

// LaunchClipClearProcess launches the timed clipboard clearing process.
// For non-interactive CLI implementations, an entirely separate process is created for this purpose.
func LaunchClipClearProcess(copySubject string, isWayland bool) {
	cmd := exec.Command(os.Args[0], "clipclear", strconv.FormatBool(isWayland))
	writeToStdin(cmd, copySubject)
	cmd.Start()
	os.Exit(0) // use os.Exit directly since this version of this function is only meant for non-interactive CLI implementations
}
