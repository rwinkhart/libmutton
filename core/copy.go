package core

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/rwinkhart/go-boilerplate/back"
	"github.com/rwinkhart/libmutton/crypt"
)

// CopyArgument copies a field from an entry to the clipboard.
func CopyArgument(targetLocation string, field int) error {
	if isFile, _ := back.TargetIsFile(targetLocation, true, 2); isFile {

		decryptedEntry, err := crypt.DecryptFileToSlice(targetLocation)
		if err != nil {
			return errors.New("unable to decrypt entry: " + err.Error())
		}
		var copySubject string // will store data to be copied

		// ensure field exists in entry
		if len(decryptedEntry) > field {

			// ensure field is not empty
			if decryptedEntry[field] == "" {
				return errors.New("field is empty")
			}

			if field != 2 {
				copySubject = decryptedEntry[field]
			} else { // TOTP mode
				var secret string // stores secret for TOTP generation
				var forSteam bool // indicates whether to generate TOTP in Steam format

				if strings.HasPrefix(decryptedEntry[2], "steam@") {
					secret = decryptedEntry[2][6:]
					forSteam = true
				} else {
					secret = decryptedEntry[2]
				}

				fmt.Println("Clipboard will be kept up to date with the current TOTP code until this process is closed")

				for { // keep token copied to clipboard, refresh on 30-second intervals
					currentTime := time.Now()
					token, err := GenTOTP(secret, currentTime, forSteam)
					if err != nil {
						return err
					}
					copyString(true, token)
					// sleep until next 30-second interval
					time.Sleep(time.Duration(30-(currentTime.Second()%30)) * time.Second)
				}
			}
		} else {
			return errors.New("field does not exist in entry")
		}

		// copy field to clipboard, launch clipboard clearing process
		copyString(false, copySubject)
	}
	return nil
}

// ClipClearArgument reads the assigned clipboard contents from stdin and passes them to clipClearProcess.
func ClipClearArgument() {
	assignedContents := back.ReadFromStdin()
	if assignedContents == "" {
		os.Exit(0) // use os.Exit instead of core.Exit, as this function runs out of a background subprocess that is invisible to the user (will never appear in GUI/TUI environment)
	}
	clipClearProcess(assignedContents)
}

// GenTOTP generates a TOTP token from a secret (supports standard and Steam TOTP).
func GenTOTP(secret string, time time.Time, forSteam bool) (string, error) {
	var totpToken string
	var err error

	if forSteam {
		totpToken, err = totp.GenerateCodeCustom(secret, time, totp.ValidateOpts{Period: 30, Digits: 5, Encoder: otp.EncoderSteam})
	} else {
		totpToken, err = totp.GenerateCode(secret, time)
	}

	if err != nil {
		return "", errors.New("unable to generate TOTP token: " + err.Error())
	}

	return totpToken, nil
}
