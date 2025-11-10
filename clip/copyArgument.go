package clip

import (
	"errors"
	"fmt"
	"os"

	"github.com/rwinkhart/go-boilerplate/back"
	"github.com/rwinkhart/libmutton/crypt"
)

// CopyArgument copies a field from an entry to the clipboard.
// If field is -1, it will one-time copy the TOTP code
// (instead of keeping the clipboard up-to-date).
func CopyArgument(targetLocation string, field int) error {
	// ensure targetLocation exists and is a file
	_, err := back.TargetIsFile(targetLocation, true)
	if err != nil {
		return err
	}

	decSlice, err := crypt.DecryptFileToSlice(targetLocation)
	if err != nil {
		return errors.New("unable to decrypt entry: " + err.Error())
	}

	// handle non-persistent TOTP copy
	var copySubject string
	var realField int
	if field == -1 {
		realField = 2
	} else {
		realField = field
	}

	// if field exists in entry...
	if len(decSlice) > realField {
		if decSlice[realField] == "" {
			return errors.New("field is empty")
		}

		if realField == 2 { // TOTP mode
			if field != -1 {
				fmt.Println("Clipboard refreshing with the current TOTP code until this process is closed")
			}
			errorChan := make(chan error)
			go TOTPCopier(decSlice[2], field, errorChan, nil) // "done" is not needed because the process runs until the program is killed
			err = <-errorChan
			if err != nil { // block until first successful copy
				return errors.New("error encountered in TOTP refresh process: " + err.Error())
			}
		} else { // other
			copySubject = decSlice[realField]
		}
	} else {
		return errors.New("field does not exist in entry")
	}

	// copy field to clipboard; launch clipboard clearing process
	err = CopyString(false, copySubject)
	if err != nil {
		return err
	}

	return nil
}

// ClipClearArgument reads the assigned clipboard contents from stdin and passes them to clipClearProcess.
func ClipClearArgument() error {
	assignedContents := back.ReadFromStdin()
	if assignedContents == "" {
		os.Exit(0) // use os.Exit instead of core.Exit, as this function runs out of a background subprocess that is invisible to the user (will never appear in GUI/TUI environment)
	}
	err := clipClearProcess(assignedContents)
	return err
}
