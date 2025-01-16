//go:build (windows || darwin || android || termux || wsl) && interactive

package core

// LaunchClipClearProcess launches the timed clipboard clearing process.
// For interactive GUI/TUI implementations, the clipboard clearing process is launched as a goroutine.
// copySubject can be omitted to clear the clipboard immediately and unconditionally.
func LaunchClipClearProcess(copySubject string) {
	go clipClearProcess(copySubject)
}
