//go:build (!android && !ios) || termux

package clip

import (
	"errors"
	"fmt"
	"os"

	"github.com/rwinkhart/go-boilerplate/back"
	"github.com/rwinkhart/libmutton/crypt"
)

// CopyShortcut (given a path) decrypts an
// entry and copies a field to the clipboard.
// Leave rcwPassword nil to use RCW demonization.
func CopyShortcut(realPath string, field int, rcwPassword []byte) error {
	// ensure realPath exists and is a file
	_, err := back.TargetIsFile(realPath, true)
	if err != nil {
		return err
	}

	// decrypt entry
	decSlice, err := crypt.DecryptFileToSlice(realPath, rcwPassword)
	if err != nil {
		return errors.New("unable to decrypt entry: " + err.Error())
	}

	// if field exists in entry...
	if len(decSlice) > field {
		if decSlice[field] == "" {
			return errors.New("field is empty")
		}

		if field == 2 { // TOTP mode
			fmt.Println(back.AnsiWarning + "[Starting]" + back.AnsiReset + " TOTP clipboard refresher")
			errorChan := make(chan error, 1)
			go TOTPCopier(decSlice[2], errorChan, nil) // "done" is not needed because the process runs until the program is killed
			if err = <-errorChan; err != nil {         // handle error from first copy
				return errors.New("error encountered in TOTP refresh process: " + err.Error())
			}
			select {} // block indefinitely
		} else { // other
			// copy field to clipboard; launch clipboard clearing process
			if err = CopyBytes(true, []byte(decSlice[field])); err != nil {
				return err
			}
			return nil
		}
	} else {
		return errors.New("field does not exist in entry")
	}
}

// ClearArgument reads the assigned clipboard contents from stdin and passes them to clipClearProcess.
func ClearArgument() error {
	assignedContents := back.ReadFromStdin()
	if assignedContents == nil {
		os.Exit(0) // use os.Exit directly since this function runs out of a background subprocess that is invisible to the user (will never appear in GUI/TUI environment)
	}
	return ClearProcess(assignedContents)
}
