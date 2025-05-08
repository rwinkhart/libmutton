package core

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

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
	launchRCWDProcess()
	return strings.Split(string(daemon.GetDec(encBytes)), "\n")
}

// EncryptBytes encrypts a byte slice using RCW and returns the encrypted data.
func EncryptBytes(decBytes []byte) []byte {
	launchRCWDProcess()
	return daemon.GetEnc(decBytes)
}

// launchRCWDProcess launches an RCW daemon to cache a passphrase.
// It returns immediately if the daemon appears to be already running.
func launchRCWDProcess() {
	if daemon.IsOpen() {
		return
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

	// block until socket file is created
	for {
		if daemon.IsOpen() {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
}
