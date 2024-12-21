//go:build (windows || darwin || android || termux || wsl) && returnOnExit

package core

// launchClipClearProcess launches the automated clipboard clearing process.
// For interactive GUI/TUI implementations, the clipboard clearing process is launched as a goroutine.
func launchClipClearProcess(copySubject string) {
	go clipClearProcess(copySubject)
}
