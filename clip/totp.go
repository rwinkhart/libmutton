//go:build (!android && !ios) || termux

package clip

import (
	"time"

	"github.com/rwinkhart/libmutton/core"
)

// TOTPCopier is meant to be run as a goroutine to keep
// the clipboard up-to-date with the latest TOTP token.
//
// Note that errorChan should be buffered with capacity 1.
// This is because TOTPCopier only returns errors on the first
// iteration, as subsequent errors are highly unlikely to occur
// and allowing the caller to move on from error-checking is beneficial.
func TOTPCopier(secret string, errorChan chan<- error, done <-chan bool) {
	var firstRun = true
	var currentTime time.Time
	var token string
	var err error
	for {
		currentTime = time.Now()
		token, err = core.GenTOTP(secret, currentTime)
		if firstRun && err != nil {
			errorChan <- err
		}
		err = CopyString(false, token)
		if firstRun {
			if err != nil {
				errorChan <- err
			} else {
				errorChan <- nil // indicate that first copy was successful
			}
			firstRun = false
		}

		// sleep till next 30-second interval
		time.Sleep(time.Duration(30-(currentTime.Second()%30)) * time.Second)

		// exit after sleep if indicated (will not update clipboard again)
		select {
		case <-done:
			return
		default:
		}
	}
}
