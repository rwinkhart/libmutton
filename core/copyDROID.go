//go:build android && !termux

package core

import (
	"golang.design/x/clipboard"
)

// copyString copies a string to the clipboard.
func copyString(continuous bool, copySubject string) {
	clipboard.Write(clipboard.FmtText, []byte(copySubject))

	if !continuous {
		LaunchClipClearProcess(copySubject)
	}
}
