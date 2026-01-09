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
var RetryPassword = true

// RCWDArgument reads the password from stdin and caches it via an RCW daemon.
func RCWDArgument() {
	password := back.ReadFromStdin()
	if password == "" {
		os.Exit(0) // use os.Exit directly since this function is only intended for non-interactive CLI clients
	}
	daemon.Start([]byte(password))
}

// DecryptFileToSlice decrypts an RCW wrapped file and returns the contents as a slice of (trimmed) strings.
func DecryptFileToSlice(realPath string) ([]string, error) {
	// read encrypted file
	encBytes, err := os.ReadFile(realPath)
	if err != nil {
		return nil, errors.New("unable to open \"" + realPath + "\" for decryption: " + err.Error())
	}

	// decrypt data using RCW daemon
	password := launchRCWDProcess()
	if password == nil {
		// if daemon is already running, use it to decrypt the data
		return strings.Split(string(daemon.GetDec(encBytes)), "\n"), nil
	}
	// if the daemon is not already running, use wrappers.Decrypt
	// directly to avoid waiting for socket file creation
	decBytes, err := wrappers.Decrypt(encBytes, password)
	if err != nil {
		return nil, errors.New("unable to decrypt \"" + realPath + "\": " + err.Error())
	}
	return strings.Split(string(decBytes), "\n"), nil
}

// EncryptBytes encrypts a byte slice using RCW and returns the encrypted data.
func EncryptBytes(decBytes []byte) []byte {
	password := launchRCWDProcess()
	if password == nil {
		// if daemon is already running, use it to encrypt the data
		return daemon.GetEnc(decBytes)
	}
	// if the daemon is not already running, use wrappers.Encrypt
	// directly to avoid waiting for socket file creation
	return wrappers.Encrypt(decBytes, password)
}

// launchRCWDProcess launches an RCW daemon to cache a password.
// If the daemon is not already running OR if not running in daemonize mode,
// it collects and returns the password (otherwise returns nil).
func launchRCWDProcess() []byte {
	if Daemonize && daemon.IsOpen() {
		return nil
	}
	var password []byte
	if RetryPassword {
		for {
			password = global.GetPassword("RCW Password:")
			if err := wrappers.RunSanityCheck(global.CfgDir+global.PathSeparator+"sanity.rcw", password); err == nil {
				break
			}
			fmt.Println(back.AnsiError + "Incorrect password" + back.AnsiReset)
		}
	} else {
		// in this mode, it is up to the client to perform the sanity check
		password = global.GetPassword("RCW Password:")
	}

	if Daemonize {
		cmd := exec.Command(os.Args[0], "startrcwd")
		cmd.SysProcAttr = global.GetSysProcAttr()
		_ = back.WriteToStdin(cmd, string(password))
		_ = cmd.Start()
	}

	return password
}
