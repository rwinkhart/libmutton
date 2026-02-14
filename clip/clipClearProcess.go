//go:build !android && !ios

package clip

import (
	"bytes"
	"errors"
	"time"

	"github.com/rwinkhart/go-boilerplate/back"
)

// ClearProcess clears the clipboard after 30 seconds if the clipboard contents have not changed.
// assignedContents can be omitted to clear the clipboard immediately and unconditionally.
func ClearProcess(assignedContents []byte) error {
	cmdPaste, cmdClear, err := getClipCommands()
	if err != nil {
		return errors.New("unable to determine clipboard platform: " + err.Error())
	}

	clearClipboard := func() error {
		if err := cmdClear.Run(); err != nil {
			return errors.New("unable to clear clipboard")
		}
		back.Exit(0)
		return nil
	}

	// if assignedContents is empty, clear the clipboard immediately and unconditionally
	if assignedContents == nil {
		if err := clearClipboard(); err != nil {
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
	if bytes.Equal(assignedContents, bytes.TrimRight(newContents, "\r\n")) {
		if err = clearClipboard(); err != nil {
			return err
		}
	}
	return nil
}
