package crypt

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/rwinkhart/go-boilerplate/back"
	"github.com/rwinkhart/libmutton/global"
	"github.com/rwinkhart/rcw/daemon"
	"github.com/rwinkhart/rcw/wrappers"
)

var Daemonize = true
var RetryPassphrase = true

// RCWDArgument reads the passphrase from stdin and caches it via an RCW daemon.
func RCWDArgument() {
	passphrase := back.ReadFromStdin()
	if passphrase == "" {
		os.Exit(0)
	}
	daemon.Start([]byte(passphrase))
}

// DecryptFileToSlice decrypts an RCW wrapped file and returns the contents as a slice of (trimmed) strings.
func DecryptFileToSlice(targetLocation string) ([]string, error) {
	// read encrypted file
	encBytes, err := os.ReadFile(targetLocation)
	if err != nil {
		return nil, errors.New("unable to open \"" + targetLocation + "\" for decryption: " + err.Error())
	}

	// decrypt data using RCW daemon
	passphrase := launchRCWDProcess()
	if passphrase == nil {
		// if daemon is already running, use it to decrypt the data
		return strings.Split(string(daemon.GetDec(encBytes)), "\n"), nil
	}
	// if the daemon is not already running, use wrappers.Decrypt
	// directly to avoid waiting for socket file creation
	decBytes, err := wrappers.Decrypt(encBytes, passphrase)
	if err != nil {
		return nil, errors.New("unable to decrypt \"" + targetLocation + "\": " + err.Error())
	}
	return strings.Split(string(decBytes), "\n"), nil
}

// EncryptBytes encrypts a byte slice using RCW and returns the encrypted data.
func EncryptBytes(decBytes []byte) []byte {
	passphrase := launchRCWDProcess()
	if passphrase == nil {
		// if daemon is already running, use it to encrypt the data
		return daemon.GetEnc(decBytes)
	}
	// if the daemon is not already running, use wrappers.Encrypt
	// directly to avoid waiting for socket file creation
	return wrappers.Encrypt(decBytes, passphrase)
}

// launchRCWDProcess launches an RCW daemon to cache a passphrase.
// If the daemon is not already running OR if not running in daemonize mode,
// it collects and returns the passphrase (otherwise returns nil).
func launchRCWDProcess() []byte {
	if Daemonize && daemon.IsOpen() {
		return nil
	}
	var passphrase []byte
	if RetryPassphrase {
		for {
			passphrase = global.GetPassphrase("RCW Passphrase:")
			err := wrappers.RunSanityCheck(global.ConfigDir+global.PathSeparator+"sanity.rcw", passphrase)
			if err == nil {
				break
			}
			fmt.Println(back.AnsiError + "Incorrect passphrase" + back.AnsiReset)
		}
	} else {
		// in this mode, it is up to the client to perform the sanity check
		passphrase = global.GetPassphrase("RCW Passphrase:")
	}

	if Daemonize {
		cmd := exec.Command(os.Args[0], "startrcwd")
		_ = back.WriteToStdin(cmd, string(passphrase))
		_ = cmd.Start()
	}

	return passphrase
}
