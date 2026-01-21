package global

import (
	"errors"
	"os"

	"github.com/rwinkhart/go-boilerplate/back"
)

// DirInit creates the libmutton directories.
// Returns: oldDeviceID (from before the directory reset; will be FSMisc if there is no pre-existing ID).
func DirInit(preserveOldCfgDir bool) (*string, error) {
	var err error

	// create EntryRoot
	if err = os.MkdirAll(EntryRoot, 0700); err != nil {
		return nil, errors.New("unable to create \"" + EntryRoot + "\": " + err.Error())
	}

	// get old device ID before its potential removal
	oldDeviceID, _ := GetCurrentDeviceID() // error ignored; oldDeviceID is set to nil on error, which is the correct assumption

	// remove existing config directory (if it exists and not in append mode)
	if !preserveOldCfgDir {
		isAccessible, _ := back.TargetIsFile(CfgDir, false) // error is ignored because dir/file status is irrelevant
		if isAccessible {
			if err = os.RemoveAll(CfgDir); err != nil {
				return nil, errors.New("unable to remove existing config directory: " + err.Error())
			}
		}
	}

	// create config directory w/devices subdirectory
	if err = os.MkdirAll(CfgDir+PathSeparator+"devices", 0700); err != nil {
		return nil, errors.New("unable to create \"" + CfgDir + "\": " + err.Error())
	}

	// create password age directory
	if err = os.MkdirAll(AgeDir, 0700); err != nil {
		return nil, errors.New("unable to create \"" + AgeDir + "\": " + err.Error())
	}

	return oldDeviceID, nil
}
