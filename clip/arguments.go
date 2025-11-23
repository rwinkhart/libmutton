//go:build (!android && !ios) || termux

package clip

import (
	"errors"
	"fmt"
	"os"

	"github.com/rwinkhart/go-boilerplate/back"
	"github.com/rwinkhart/libmutton/crypt"
)

// CopyShortcut, given a path, decrypts an
// entry and copies a field to the clipboard.
func CopyShortcut(realPath string, field int) error {
	// ensure realPath exists and is a file
	_, err := back.TargetIsFile(realPath, true)
	if err != nil {
		return err
	}

	// decrypt entry
	decSlice, err := crypt.DecryptFileToSlice(realPath)
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
			errorChan := make(chan error)
			go TOTPCopier(decSlice[2], errorChan, nil) // "done" is not needed because the process runs until the program is killed
			err = <-errorChan
			if err != nil { // handle error from first copy
				return errors.New("error encountered in TOTP refresh process: " + err.Error())
			}
			if field != -1 {
				fmt.Println(back.AnsiGreen + "[Started]" + back.AnsiReset + " TOTP clipboard refresher\n\nService will run until this process is killed")
			}
			select {} // block indefinitely
		} else { // other
			// copy field to clipboard; launch clipboard clearing process
			err = CopyString(true, decSlice[field])
			if err != nil {
				return err
			}
			return nil
		}
	} else {
		return errors.New("field does not exist in entry")
	}
}

// ClipClearArgument reads the assigned clipboard contents from stdin and passes them to clipClearProcess.
func ClipClearArgument() error {
	assignedContents := back.ReadFromStdin()
	if assignedContents == "" {
		os.Exit(0) // use os.Exit instead of core.Exit, as this function runs out of a background subprocess that is invisible to the user (will never appear in GUI/TUI environment)
	}
	err := ClipClearProcess(assignedContents)
	return err
}
