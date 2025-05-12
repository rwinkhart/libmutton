package global

import (
	"io/fs"
	"os"

	"github.com/rwinkhart/go-boilerplate/back"
)

// GetOldDeviceID returns the current device ID or
// FSMisc if there is no device ID (e.g. first run).
func GetCurrentDeviceID() string {
	deviceIDList := GenDeviceIDList(false) // errorOnFail is false so that nil is received when the devices directory does not exist
	var deviceID string
	if len(deviceIDList) > 0 {
		deviceID = (deviceIDList)[0].Name()
	} else {
		deviceID = FSMisc // indicates to server that no device ID is being replaced
	}
	return deviceID
}

// GenDeviceIDList returns a slice of all registered device IDs.
// Requires: errorOnFail (set to true to throw an error if the devices directory cannot be read/does not exist)
func GenDeviceIDList(errorOnFail bool) []fs.DirEntry {
	// create a slice of all registered devices
	deviceIDList, err := os.ReadDir(ConfigDir + PathSeparator + "devices")
	if err != nil {
		if errorOnFail {
			back.PrintError("Failed to read the devices directory: "+err.Error(), back.ErrorRead, true)
		} else {
			return nil // a nil return value indicates that the devices directory could not be read/does not exist
		}
	}
	return deviceIDList
}
