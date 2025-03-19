//go:build android && !termux

package core

import (
	"golang.design/x/clipboard"
)

// TODO Investigate background clipboard clearing and on-app-close clipboard clearing for Android

// copyString copies a string to the clipboard.
func copyString(continuous bool, copySubject string) {
	clipboard.Write(clipboard.FmtText, []byte(copySubject))

	if !continuous {
		LaunchClipClearProcess(copySubject)
	}
}
