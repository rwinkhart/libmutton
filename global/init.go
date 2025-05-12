package global

import (
	"os"

	"github.com/rwinkhart/go-boilerplate/back"
)

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
