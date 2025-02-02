package sync

import (
	"os"
	"strings"

	"github.com/rwinkhart/libmutton/core"
)

// getModTimes returns a list of all entry modification times.
func getModTimes(entryList []string) []int64 {
	var modList []int64
	for _, file := range entryList {
		modTime, _ := os.Stat(core.TargetLocationFormat(file))
		modList = append(modList, modTime.ModTime().Unix())
	}

	return modList
}

// ShearLocal removes the target file or directory from the local system.
// Returns: deviceID (only on client; for use in ShearRemoteFromClient),
// isDir (only on client; for use in ShearRemoteFromClient).
// If the local system is a server, it will also add the target to the deletions list for all clients (except the requesting client).
// This function should only be used directly by the server binary.
func ShearLocal(targetLocationIncomplete, clientDeviceID string) (string, bool) {
	// determine if running on a server
	var onServer bool
	if clientDeviceID != "" {
		onServer = true
	}

	deviceIDList := core.GenDeviceIDList(true)

	// add the sheared target (incomplete, vanity) to the deletions list (if running on a server)
	if onServer {
		for _, device := range *deviceIDList {
			if device.Name() != clientDeviceID {
				fileToClose, err := os.OpenFile(core.ConfigDir+core.PathSeparator+"deletions"+core.PathSeparator+device.Name()+core.FSSpace+strings.ReplaceAll(targetLocationIncomplete, "/", core.FSPath), os.O_CREATE|os.O_WRONLY, 0600)
				if err != nil {
					// do not print error as there is currently no way of seeing server-side errors
					// failure to add the target to the deletions list will exit the program and result in a client re-uploading the target (non-critical)
					os.Exit(core.ErrorWrite)
				}
				_ = fileToClose.Close() // error ignored; if the file could be created, it can probably be closed
			}
		}
	}

	// get the full targetLocation path and remove the target
	targetLocationComplete := core.TargetLocationFormat(targetLocationIncomplete)
	var isFile bool
	if !onServer { // error if target does not exist on client, needed because os.RemoveAll does not return an error if target does not exist
		isFile, _ = core.TargetIsFile(targetLocationComplete, true, 0)
	}
	err := os.RemoveAll(targetLocationComplete)
	if err != nil {
		core.PrintError("Failed to remove local target: "+err.Error(), core.ErrorWrite, true)
	}

	if !onServer && len(*deviceIDList) > 0 { // return the device ID if running on the client and a device ID exists (online mode)
		return (*deviceIDList)[0].Name(), !isFile
	}
	return "", true

	// do not exit program, as this function is used as part of ShearRemoteFromClient
}

// RenameLocal renames oldLocationIncomplete to newLocationIncomplete on the local system.
// This function should only be used directly by the server binary.
func RenameLocal(oldLocationIncomplete, newLocationIncomplete string, verifyOldLocationExists bool) {
	// get full paths for both locations
	oldLocation := core.TargetLocationFormat(oldLocationIncomplete)
	newLocation := core.TargetLocationFormat(newLocationIncomplete)

	if verifyOldLocationExists {
		core.TargetIsFile(oldLocation, true, 0)
	}

	// ensure newLocation does not exist
	_, isAccessible := core.TargetIsFile(newLocation, false, 0)
	if isAccessible {
		core.PrintError("\""+newLocation+"\" already exists", core.ErrorTargetExists, true)
	}

	// rename oldLocation to newLocation
	err := os.Rename(oldLocation, newLocation)
	if err != nil {
		core.PrintError("Failed to rename - Does the target containing directory exist?", core.ErrorTargetNotFound, true)
	}

	// do not exit program, as this function is used as part of RenameRemoteFromClient
}

// AddFolderLocal creates a new entry-containing directory on the local system.
// This function should only be used directly by the server binary.
func AddFolderLocal(targetLocationIncomplete string) {
	// get the full targetLocation path and create the target
	targetLocationComplete := core.TargetLocationFormat(targetLocationIncomplete)
	err := os.Mkdir(targetLocationComplete, 0700)
	if err != nil {
		if os.IsExist(err) {
			core.PrintError("Directory already exists", core.ErrorTargetExists, true)
		} else {
			core.PrintError("Failed to create directory: "+err.Error(), core.ErrorWrite, true)
		}
	}

	// do not exit program, as this function is used as part of AddFolderRemoteFromClient
}
