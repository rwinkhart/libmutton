package core

import (
	"fmt"
	"os"
	"strings"
	"time"

	steamtotp "github.com/fortis/go-steam-totp"
	"github.com/pquerna/otp/totp"
)

// CopyArgument copies a field from an entry to the clipboard.
func CopyArgument(targetLocation string, field int, passphrase []byte) {
	if isFile, _ := TargetIsFile(targetLocation, true, 2); isFile {

		decryptedEntry := DecryptFileToSlice(targetLocation, passphrase)
		var copySubject string // will store data to be copied

		// ensure field exists in entry
		if len(decryptedEntry) > field {

			// ensure field is not empty
			if decryptedEntry[field] == "" {
				PrintError("Field is empty", ErrorTargetNotFound, true)
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
					copyString(true, GenTOTP(secret, currentTime, forSteam))
					// sleep until next 30-second interval
					time.Sleep(time.Duration(30-(currentTime.Second()%30)) * time.Second)
				}
			}
		} else {
			PrintError("Field does not exist in entry", ErrorTargetNotFound, true)
		}

		// copy field to clipboard, launch clipboard clearing process
		copyString(false, copySubject)
	}
}

// ClipClearArgument reads the assigned clipboard contents from stdin and passes them to clipClearProcess.
func ClipClearArgument() {
	assignedContents := readFromStdin()
	if assignedContents == "" {
		os.Exit(0) // use os.Exit instead of core.Exit, as this function runs out of a background subprocess that is invisible to the user (will never appear in GUI/TUI environment)
	}
	clipClearProcess(assignedContents)
}

// GenTOTP generates a TOTP token from a secret (supports standard and Steam TOTP).
func GenTOTP(secret string, time time.Time, forSteam bool) string {
	var totpToken string
	var err error

	if forSteam {
		totpToken, err = steamtotp.GenerateAuthCode(secret, time)
	} else {
		totpToken, err = totp.GenerateCode(secret, time)
	}

	if err != nil {
		PrintError("Error generating TOTP code", ErrorOther, true)
	}

	return totpToken
}
