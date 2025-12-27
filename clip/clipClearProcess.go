//go:build (!android && !ios) || termux

package clip

import (
	"errors"
	"strings"
	"time"

	"github.com/rwinkhart/go-boilerplate/back"
)

// ClearProcess clears the clipboard after 30 seconds if the clipboard contents have not changed.
// assignedContents can be omitted to clear the clipboard immediately and unconditionally.
func ClearProcess(assignedContents string) error {
	cmdPaste, cmdClear, err := getClipCommands()
	if err != nil {
		return errors.New("unable to determine clipboard platform: " + err.Error())
	}

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
		if err = clearClipboard(); err != nil {
			return err
		}
	}
	return nil
}
