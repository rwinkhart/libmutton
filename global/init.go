package global

import (
	"errors"
	"os"

	"github.com/rwinkhart/go-boilerplate/back"
)

// DirInit creates the libmutton directories.
// Returns: oldDeviceID (from before the directory reset; will be FSMisc if there is no pre-existing ID).
func DirInit(preserveOldConfigDir bool) (string, error) {
	// create EntryRoot
	err := os.MkdirAll(EntryRoot, 0700)
	if err != nil {
		return "", errors.New("unable to create \"" + EntryRoot + "\": " + err.Error())
	}

	// get old device ID before its potential removal
	oldDeviceID, err := GetCurrentDeviceID()
	if err != nil {
		oldDeviceID = FSMisc
	}

	// remove existing config directory (if it exists and not in append mode)
	if !preserveOldConfigDir {
		isAccessible, _ := back.TargetIsFile(ConfigDir, false) // error is ignored because dir/file status is irrelevant
		if isAccessible {
			err = os.RemoveAll(ConfigDir)
			if err != nil {
				return "", errors.New("unable to remove existing config directory: " + err.Error())
			}
		}
	}

	// create config directory w/devices subdirectory
	err = os.MkdirAll(ConfigDir+PathSeparator+"devices", 0700)
	if err != nil {
		return "", errors.New("unable to create \"" + ConfigDir + "\": " + err.Error())
	}

	return oldDeviceID, nil
}
