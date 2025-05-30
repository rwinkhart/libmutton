//go:build !android || termux

package core

import (
	"errors"
	"strings"
	"time"

	"github.com/rwinkhart/go-boilerplate/back"
)

// clipClearProcess clears the clipboard after 30 seconds if the clipboard contents have not changed.
// assignedContents can be omitted to clear the clipboard immediately and unconditionally.
func clipClearProcess(assignedContents string) error {
	cmdPaste, cmdClear := getClipCommands()

	clearClipboard := func() error {
		err := cmdClear.Run()
		if err != nil {
			return errors.New("unable to clear clipboard")
		}
		back.Exit(0)
		return nil
	}

	// if assignedContents is empty, clear the clipboard immediately and unconditionally
	if assignedContents == "" {
		err := clearClipboard()
		if err != nil {
			return err
		}
		return nil
	}

	// wait 30 seconds before checking clipboard contents
	time.Sleep(30 * time.Second)

	newContents, err := cmdPaste.Output()
	if err != nil {
		return errors.New("unable to read clipboard contents")
	}

	if assignedContents == strings.TrimRight(string(newContents), "\r\n") {
		err := clearClipboard()
		if err != nil {
			return err
		}
	}
	return nil
}
