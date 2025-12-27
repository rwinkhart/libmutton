//go:build !ios && (!android || termux) && interactive

package clip

// LaunchClearProcess launches the timed clipboard clearing process.
// For interactive GUI/TUI implementations, the clipboard clearing process is launched as a goroutine.
// copySubject can be omitted to clear the clipboard immediately and unconditionally.
func LaunchClearProcess(copySubject string) {
	go ClearProcess(copySubject)
}
