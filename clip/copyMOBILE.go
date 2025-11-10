//go:build (android && !termux) || ios

package clip

import (
	"golang.design/x/clipboard"
)

// CopyString copies a string to the clipboard.
func CopyString(continuous bool, copySubject string) error {
	clipboard.Write(clipboard.FmtText, []byte(copySubject))
	if !continuous {
		LaunchClipClearProcess(copySubject)
	}
	return nil
}
