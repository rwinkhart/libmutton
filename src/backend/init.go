package backend

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// TempInit ensures libmutton directories exist and writes the libmutton configuration file
func TempInit(configFileMap map[string]string) {
	// create EntryRoot and ConfigDir
	dirInit()

	// remove existing config file
	removeFile(ConfigPath)

	if configFileMap["textEditor"] == "" {
		configFileMap["textEditor"] = textEditorFallback()
	}

	// create and write config file
	configFile, _ := os.OpenFile(ConfigPath, os.O_CREATE|os.O_WRONLY, 0600)
	defer configFile.Close()
	configFile.WriteString("[LIBMUTTON]\n")
	for key, value := range configFileMap {
		configFile.WriteString(key + " = " + value + "\n")
	}

	os.Exit(0)
}

// GpgUIDListGen generates a list of all GPG key IDs on the system and returns them as a slice of strings
func GpgUIDListGen() []string {
	cmd := exec.Command("gpg", "-k", "--with-colons")
	gpgOutputBytes, _ := cmd.Output()
	gpgOutputLines := strings.Split(string(gpgOutputBytes), "\n")
	var uidSlice []string
	for _, line := range gpgOutputLines {
		if strings.HasPrefix(line, "uid") {
			uid := strings.Split(line, ":")[9]
			uidSlice = append(uidSlice, uid)
		}
	}
	return uidSlice
}

// GpgKeyGen generates a new GPG key and returns the key ID
func GpgKeyGen() string {
	gpgGenTempFile := CreateTempFile()
	defer os.Remove(gpgGenTempFile.Name())

	// create and write gpg-gen file
	unixTime := strconv.FormatInt(time.Now().Unix(), 10)
	gpgGenTempFile.WriteString(strings.Join([]string{"Key-Type: eddsa", "Key-Curve: ed25519", "Key-Usage: sign", "Subkey-Type: ecdh", "Subkey-Curve: cv25519", "Subkey-Usage: encrypt", "Name-Real: libmutton-" + unixTime, "Name-Comment: gpg-libmutton", "Name-Email: github.com/rwinkhart/libmutton", "Expire-Date: 0"}, "\n"))

	// close gpg-gen file
	gpgGenTempFile.Close()

	// generate GPG key based on gpg-gen file
	cmd := exec.Command("gpg", "-q", "--batch", "--generate-key", gpgGenTempFile.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Run()

	return "libmutton-" + unixTime + " (gpg-libmutton) <github.com/rwinkhart/libmutton>"
}