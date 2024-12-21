//go:build returnOnExit

package core

// launchClipClear launches the automated clipboard clearing process.
// For interactive GUI/TUI implementations, the clipboard clearing process is launched as a goroutine.
func launchClipClear(copySubject string) {
	go clipClear(copySubject)
}
