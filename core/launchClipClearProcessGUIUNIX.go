//go:build !windows && !darwin && !android && !termux && !wsl && returnOnExit

package core

import (
	"os"
	"strconv"
)

// launchClipClearProcess launches the automated clipboard clearing process.
// For interactive GUI/TUI implementations, the clipboard clearing process is launched as a goroutine.
func launchClipClearProcess(copySubject string, isWayland bool) {
	os.Args[2] = strconv.FormatBool(isWayland)
	go clipClearProcess(copySubject)
}
