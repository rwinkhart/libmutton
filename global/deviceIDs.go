package global

import (
	"errors"
	"io/fs"
	"os"
)

// GetCurrentDeviceID returns the current device ID or
// FSMisc if there is no device ID (e.g. first run).
func GetCurrentDeviceID() (string, error) {
	deviceIDList, err := GenDeviceIDList()
	if err != nil {
		return "", errors.New("unable to generate device ID list: " + err.Error())
	}
	var deviceID string
	if len(deviceIDList) > 0 {
		deviceID = (deviceIDList)[0].Name()
	} else {
		deviceID = FSMisc // indicates to server that no device ID is being replaced
	}
	return deviceID, nil
}

// GenDeviceIDList returns a slice of all registered device IDs.
// Requires: errorOnFail (set to true to throw an error if the devices directory cannot be read/does not exist)
func GenDeviceIDList() ([]fs.DirEntry, error) {
	// create a slice of all registered devices
	deviceIDList, err := os.ReadDir(CfgDir + PathSeparator + "devices")
	if err != nil {
		return nil, errors.New("unable to read devices directory: " + err.Error())
	}
	return deviceIDList, nil
}
