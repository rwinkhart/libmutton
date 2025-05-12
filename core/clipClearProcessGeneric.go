//go:build !android || termux

package core

import (
	"strings"
	"time"

	"github.com/rwinkhart/go-boilerplate/back"
	"github.com/rwinkhart/libmutton/global"
)

// clipClearProcess clears the clipboard after 30 seconds if the clipboard contents have not changed.
// assignedContents can be omitted to clear the clipboard immediately and unconditionally.
func clipClearProcess(assignedContents string) {
	cmdPaste, cmdClear := getClipCommands()

	clearClipboard := func() {
		err := cmdClear.Run()
		if err != nil {
			back.PrintError("Failed to clear clipboard", global.ErrorClipboard, true)
		}
		back.Exit(0)
	}

	// if assignedContents is empty, clear the clipboard immediately and unconditionally
	if assignedContents == "" {
		clearClipboard()
		return
	}

	// wait 30 seconds before checking clipboard contents
	time.Sleep(30 * time.Second)

	newContents, err := cmdPaste.Output()
	if err != nil {
		back.PrintError("Failed to read clipboard contents", global.ErrorClipboard, true)
	}

	if assignedContents == strings.TrimRight(string(newContents), "\r\n") {
		clearClipboard()
	}
}
