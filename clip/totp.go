package clip

import (
	"errors"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

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

// TOTPCopier is meant to be run as a goroutine to keep
// the clipboard up-to-date with the latest TOTP token.
// Set oneTime to -1 for a one-time (non-continuous) TOTP copy.
func TOTPCopier(secret string, oneTime int, errorChan chan<- error, done <-chan bool) {
	var forSteam bool
	if strings.HasPrefix(secret, "steam@") {
		secret = secret[6:]
		forSteam = true
	}

	for {
		currentTime := time.Now()
		token, err := GenTOTP(secret, currentTime, forSteam)
		if err != nil {
			errorChan <- err
		}
		err = CopyString(true, token)
		if err != nil {
			errorChan <- err
		}
		if oneTime != -1 {
			errorChan <- nil
			time.Sleep(time.Duration(30-(currentTime.Second()%30)) * time.Second)
		} else {
			return
		}

		// exit after sleep if indicated (will not update clipboard again)
		select {
		case <-done:
			errorChan <- nil
			return
		default:
		}
	}
}
