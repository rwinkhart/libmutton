//go:build !windows && !darwin && !android && !ios && !termux && !wsl && interactive

package clip

import (
	"os"
	"strconv"
)

// LaunchClipClearProcess launches the timed clipboard clearing process.
// For interactive GUI/TUI implementations, the clipboard clearing process is launched as a goroutine.
// copySubject can be omitted to clear the clipboard immediately and unconditionally.
func LaunchClipClearProcess(copySubject string, isWayland bool) {
	os.Args = []string{os.Args[0], "", strconv.FormatBool(isWayland)}
	go clipClearProcess(copySubject)
}
