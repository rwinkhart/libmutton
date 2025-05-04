package core

import (
	"os"
	"os/exec"
	"strings"

	"github.com/rwinkhart/rcw/daemon"
	"github.com/rwinkhart/rcw/wrappers"
)

// DecryptFileToSlice decrypts an RCW wrapped file and returns the contents as a slice of (trimmed) strings.
func DecryptFileToSlice(targetLocation string, passphrase []byte) []string {
	encBytes, err := os.ReadFile(targetLocation)
	if err != nil {
		PrintError("Failed to decrypt \""+targetLocation+"\" - "+err.Error(), ErrorDecryption, true)
	}
	decBytes, err := wrappers.Decrypt(encBytes, passphrase)
	if err != nil {
		PrintError("Failed to decrypt \""+targetLocation+"\" - "+err.Error(), ErrorDecryption, true)
	}
	return strings.Split(string(decBytes), "\n")
}

// EncryptBytes encrypts a byte slice using RCW and returns the encrypted data.
func EncryptBytes(decBytes []byte, passphrase []byte) []byte {
	encBytes := wrappers.Encrypt(decBytes, passphrase)
	return encBytes
}

// LaunchRCWDProcess launches an RCW daemon to serve the given passphrase.
func LaunchRCWDProcess(passphrase string) {
	cmd := exec.Command(os.Args[0], "startrcwd")
	writeToStdin(cmd, passphrase)
	err := cmd.Start()
	if err != nil {
		PrintError("Failed to launch RCW daemon - Does this libmutton implementation support the \"startrcwd\" argument?", ErrorOther, true)
	}
}

// RCWDArgument reads the passphrase from stdin and serves it via an RCW daemon.
func RCWDArgument() {
	passphrase := readFromStdin()
	if passphrase == "" {
		os.Exit(0)
	}
	daemon.Start(string(passphrase))
}
