package core

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	steamtotp "github.com/fortis/go-steam-totp"
	"github.com/pquerna/otp/totp"
)

// CopyArgument copies a field from an entry to the clipboard.
func CopyArgument(targetLocation string, field int) {
	if isFile, _ := TargetIsFile(targetLocation, true, 2); isFile {

		decryptedEntry := DecryptGPG(targetLocation)
		var copySubject string // will store data to be copied

		// ensure field exists in entry
		if len(decryptedEntry) > field {

			// ensure field is not empty
			if decryptedEntry[field] == "" {
				fmt.Println(AnsiError + "Field is empty" + AnsiReset)
				os.Exit(ErrorTargetNotFound)
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
			fmt.Println(AnsiError + "Field does not exist in entry" + AnsiReset)
			os.Exit(ErrorTargetNotFound)
		}

		// copy field to clipboard, launch clipboard clearing process
		copyString(false, copySubject)
	}
}

// ClipClearArgument reads the assigned clipboard contents from stdin and passes them to clipClearProcess.
func ClipClearArgument() {
	// read previous clipboard contents from stdin
	clipScanner := bufio.NewScanner(os.Stdin)
	if clipScanner.Scan() {
		assignedContents := clipScanner.Text()
		clipClearProcess(assignedContents)
	} else {
		os.Exit(0) // use os.Exit instead of core.Exit, as this function runs out of a background subprocess that is invisible to the user (will never appear in GUI/TUI environment)
	}
}

// clipClearProcess clears the clipboard after 30 seconds if the clipboard contents have not changed.
// assignedContents can be omitted to clear the clipboard immediately and unconditionally.
func clipClearProcess(assignedContents string) {
	cmdPaste, cmdClear := getClipCommands()

	clearClipboard := func() {
		err := cmdClear.Run()
		if err != nil {
			fmt.Println(AnsiError+"Failed to clear clipboard:", err.Error()+AnsiReset)
			os.Exit(ErrorClipboard)
		}
		Exit(0)
	}

	// if assignedContents is empty, clear the clipboard immediately and unconditionally
	if assignedContents == "" {
		clearClipboard()
	}

	// wait 30 seconds before checking clipboard contents
	time.Sleep(30 * time.Second)

	newContents, err := cmdPaste.Output()
	if err != nil {
		fmt.Println(AnsiError+"Failed to read clipboard contents:", err.Error()+AnsiReset)
		os.Exit(ErrorClipboard)
	}

	if assignedContents == strings.TrimRight(string(newContents), "\r\n") {
		clearClipboard()
	}
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
		fmt.Println(AnsiError + "Error generating TOTP code" + AnsiReset)
		os.Exit(ErrorOther)
	}

	return totpToken
}
