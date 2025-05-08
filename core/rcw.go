package core

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/rwinkhart/rcw/daemon"
	"github.com/rwinkhart/rcw/wrappers"
)

// RCWDArgument reads the passphrase from stdin and caches it via an RCW daemon.
func RCWDArgument() {
	passphrase := readFromStdin()
	if passphrase == "" {
		os.Exit(0)
	}
	daemon.Start([]byte(passphrase))
}

// DecryptFileToSlice decrypts an RCW wrapped file and returns the contents as a slice of (trimmed) strings.
func DecryptFileToSlice(targetLocation string) []string {
	// read encrypted file
	encBytes, err := os.ReadFile(targetLocation)
	if err != nil {
		PrintError("Failed to decrypt \""+targetLocation+"\" - "+err.Error(), ErrorDecryption, true)
	}

	// decrypt data using RCW daemon
	passphrase := launchRCWDProcess()
	if passphrase == nil {
		// if daemon is already running, use it to decrypt the data
		return strings.Split(string(daemon.GetDec(encBytes)), "\n")
	}
	// if the daemon is not already running, use wrappers.Decrypt
	// directly to avoid waiting for socket file creation
	decBytes, err := wrappers.Decrypt(encBytes, passphrase)
	return strings.Split(string(decBytes), "\n")
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
// If the daemon is not already running, it returns the passphrase (otherwise returns nil).
func launchRCWDProcess() []byte {
	if daemon.IsOpen() {
		return nil
	}
	var passphrase []byte
	for {
		passphrase = GetPassphrase()
		err := wrappers.RunSanityCheck(ConfigDir+PathSeparator+"sanity.rcw", passphrase)
		if err == nil {
			break
		}
		fmt.Println(AnsiError + "Incorrect passphrase" + AnsiReset)
	}
	cmd := exec.Command(os.Args[0], "startrcwd")
	writeToStdin(cmd, string(passphrase))
	cmd.Start()

	return passphrase
}
