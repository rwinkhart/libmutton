//go:build (!android && !ios) || termux

package clip

import (
	"time"

	"github.com/rwinkhart/libmutton/core"
)

// TOTPCopier is meant to be run as a goroutine to keep
// the clipboard up-to-date with the latest TOTP token.
func TOTPCopier(secret string, errorChan chan<- error, done <-chan bool) {
	for {
		currentTime := time.Now()
		token, err := core.GenTOTP(secret, currentTime)
		if err != nil {
			errorChan <- err
		}
		err = CopyString(false, token)
		if err != nil {
			errorChan <- err
		}

		errorChan <- nil // indicate that first copy was successful

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
