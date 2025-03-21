package core

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// GpgUIDListGen generates a list of all GPG key IDs on the system and returns them as a slice of strings.
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

// GpgKeyGen generates a new GPG key and returns the key ID.
func GpgKeyGen() string {
	gpgGenTempFile := CreateTempFile()
	defer func(name string) {
		_ = os.Remove(name) // error ignored; if the file could be created, it can probably be removed
	}(gpgGenTempFile.Name())

	// create and write gpg-gen file
	unixTime := strconv.FormatInt(time.Now().Unix(), 10)
	_, _ = gpgGenTempFile.WriteString(strings.Join([]string{"Key-Type: eddsa", "Key-Curve: ed25519", "Key-Usage: sign", "Subkey-Type: ecdh", "Subkey-Curve: cv25519", "Subkey-Usage: encrypt", "Name-Real: libmutton-" + unixTime, "Name-Comment: gpg-libmutton", "Name-Email: github.com/rwinkhart/libmutton", "Expire-Date: 0"}, "\n")) // error ignored; if the file could be created, it can probably be written to

	// close gpg-gen file
	_ = gpgGenTempFile.Close() // error ignored; if the file could be created, it can probably be closed

	// generate GPG key based on gpg-gen file
	cmd := exec.Command("gpg", "-q", "--batch", "--generate-key", gpgGenTempFile.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if err != nil {
		PrintError("Failed to generate GPG key: "+err.Error(), ErrorOther, true)
	}

	return "libmutton-" + unixTime + " (gpg-libmutton) <github.com/rwinkhart/libmutton>"
}

// DirInit creates the libmutton directories.
// Returns: oldDeviceID (from before the directory reset; will be FSMisc if there is no pre-existing ID).
func DirInit(preserveOldConfigDir bool) string {
	// create EntryRoot
	err := os.MkdirAll(EntryRoot, 0700)
	if err != nil {
		PrintError("Failed to create \""+EntryRoot+"\": "+err.Error(), ErrorWrite, true)
	}

	// get old device ID before its potential removal
	oldDeviceID := GetCurrentDeviceID()

	// remove existing config directory (if it exists and not in append mode)
	if !preserveOldConfigDir {
		_, isAccessible := TargetIsFile(ConfigDir, false, 1)
		if isAccessible {
			err = os.RemoveAll(ConfigDir)
			if err != nil {
				PrintError("Failed to remove existing config directory: "+err.Error(), ErrorWrite, true)
			}
		}
	}

	// create config directory w/devices subdirectory
	err = os.MkdirAll(ConfigDir+PathSeparator+"devices", 0700)
	if err != nil {
		PrintError("Failed to create \""+ConfigDir+"\": "+err.Error(), ErrorWrite, true)
	}

	return oldDeviceID
}

// GetOldDeviceID returns the current device ID or
// FSMisc if there is no device ID (e.g. first run).
func GetCurrentDeviceID() string {
	deviceIDList := GenDeviceIDList(false) // errorOnFail is false so that nil is received when the devices directory does not exist
	var deviceID string
	if deviceIDList != nil && len(*deviceIDList) > 0 { // ensure not derferencing nil, which occurs when the devices directory does not exist
		deviceID = (*deviceIDList)[0].Name()
	} else {
		deviceID = FSMisc // indicates to server that no device ID is being replaced
	}
	return deviceID
}
