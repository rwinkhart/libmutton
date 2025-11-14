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
// For oneTime mode, this function can be used normally (not as a goroutine).
func TOTPCopier(secret string, oneTime int, errorChan chan<- error, done <-chan bool) error {
	var writeErrorChan func(err error)
	if errorChan != nil {
		writeErrorChan = func(err error) { errorChan <- err }
	} else {
		writeErrorChan = func(err error) {}
	}

	var forSteam bool
	if strings.HasPrefix(secret, "steam@") {
		secret = secret[6:]
		forSteam = true
	}

	for {
		currentTime := time.Now()
		token, err := GenTOTP(secret, currentTime, forSteam)
		if err != nil {
			writeErrorChan(err)
			return err // return for when not used as goroutine; should exit on error regardless
		}
		err = CopyString(true, token)
		if err != nil {
			writeErrorChan(err)
			return err // return for when not used as goroutine; should exit on error regardless
		}
		if oneTime != -1 {
			writeErrorChan(err)
			time.Sleep(time.Duration(30-(currentTime.Second()%30)) * time.Second)
		} else {
			return nil
		}

		// exit after sleep if indicated (will not update clipboard again)
		select {
		case <-done:
			writeErrorChan(nil)
			return nil
		default:
		}
	}
}
