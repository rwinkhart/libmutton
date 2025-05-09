package core

import (
	"os"

	"github.com/rwinkhart/go-boilerplate/back"
	"github.com/rwinkhart/rcw/wrappers"
)

// RCWSanityCheckGen generates the RCW sanity check file for libmutton.
func RCWSanityCheckGen(passphrase []byte) {
	err := wrappers.GenSanityCheck(ConfigDir+PathSeparator+"sanity.rcw", passphrase)
	if err != nil {
		back.PrintError("Failed to generate sanity check file: "+err.Error(), back.ErrorWrite, true)
	}
}

// DirInit creates the libmutton directories.
// Returns: oldDeviceID (from before the directory reset; will be FSMisc if there is no pre-existing ID).
func DirInit(preserveOldConfigDir bool) string {
	// create EntryRoot
	err := os.MkdirAll(EntryRoot, 0700)
	if err != nil {
		back.PrintError("Failed to create \""+EntryRoot+"\": "+err.Error(), back.ErrorWrite, true)
	}

	// get old device ID before its potential removal
	oldDeviceID := GetCurrentDeviceID()

	// remove existing config directory (if it exists and not in append mode)
	if !preserveOldConfigDir {
		_, isAccessible := back.TargetIsFile(ConfigDir, false, 1)
		if isAccessible {
			err = os.RemoveAll(ConfigDir)
			if err != nil {
				back.PrintError("Failed to remove existing config directory: "+err.Error(), back.ErrorWrite, true)
			}
		}
	}

	// create config directory w/devices subdirectory
	err = os.MkdirAll(ConfigDir+PathSeparator+"devices", 0700)
	if err != nil {
		back.PrintError("Failed to create \""+ConfigDir+"\": "+err.Error(), back.ErrorWrite, true)
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
