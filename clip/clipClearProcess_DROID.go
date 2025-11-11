//go:build (android && !termux) || ios

package clip

import (
	"strings"
	"time"

	"github.com/rwinkhart/go-boilerplate/back"
	"golang.design/x/clipboard"
)

// ClipClearProcess clears the clipboard after 30 seconds if the clipboard contents have not changed.
// assignedContents can be omitted to clear the clipboard immediately and unconditionally.
func ClipClearProcess(assignedContents string) error {
	clearClipboard := func() {
		clipboard.Write(clipboard.FmtText, []byte(""))
		back.Exit(0)
	}

	// if assignedContents is empty, clear the clipboard immediately and unconditionally
	if assignedContents == "" {
		clearClipboard()
		return nil
	}

	// wait 30 seconds before checking clipboard contents
	time.Sleep(30 * time.Second)

	newContents := clipboard.Read(clipboard.FmtText)

	if assignedContents == strings.TrimRight(string(newContents), "\r\n") {
		clearClipboard()
	}
	return nil
}
