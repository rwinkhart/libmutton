//go:build (!android && !ios) || termux

package clip

import (
	"errors"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// GenTOTP generates a TOTP token from a secret.
// Prefix `secret` with "steam@" for Steam TOTP format.
func GenTOTP(secret string, time time.Time) (string, error) {
	var totpToken string
	var err error

	if strings.HasPrefix(secret, "steam@") {
		totpToken, err = totp.GenerateCodeCustom(secret[6:], time, totp.ValidateOpts{Period: 30, Digits: 5, Encoder: otp.EncoderSteam})
	} else {
		totpToken, err = totp.GenerateCode(secret, time)
	}

	if err != nil {
		return "", errors.New("unable to generate TOTP token: " + err.Error())
	}

	return totpToken, nil
}

// TOTPCopier is meant to be run as a goroutine to keep
// the clipboard up-to-date with the latest TOTP token.
func TOTPCopier(secret string, errorChan chan<- error, done <-chan bool) {
	for {
		currentTime := time.Now()
		token, err := GenTOTP(secret, currentTime)
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
