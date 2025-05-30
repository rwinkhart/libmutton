package synccommon

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/rwinkhart/go-boilerplate/back"
	"github.com/rwinkhart/libmutton/global"
)

// ANSI color constants used only in this file
const (
	AnsiDelete   = "\033[38;5;1m"
	AnsiDownload = "\033[38;5;2m"
	AnsiUpload   = "\033[38;5;4m"
)

var RootLength = len(global.EntryRoot) // length of global.EntryRoot string

// GetModTimes returns a list of all entry modification times.
func GetModTimes(entryList []string) []int64 {
	var modList []int64
	for _, file := range entryList {
		modTime, _ := os.Stat(global.TargetLocationFormat(file))
		modList = append(modList, modTime.ModTime().Unix())
	}

	return modList
}

// ShearLocal removes the target file or directory from the local system.
// Returns: deviceID (only on client; for use in ShearRemoteFromClient),
// isDir (only on client; for use in ShearRemoteFromClient).
// If the local system is a server, it will also add the target to the deletions list for all clients (except the requesting client).
// This function should only be used directly by the server binary.
func ShearLocal(targetLocationIncomplete, clientDeviceID string) (string, bool, error) {
	// determine if running on a server
	var onServer bool
	if clientDeviceID != "" {
		onServer = true
	}

	deviceIDList, err := global.GenDeviceIDList()
	if err != nil {
		return "", false, errors.New("unable to generate device ID list: " + err.Error())
	}

	// add the sheared target (incomplete, vanity) to the deletions list (if running on a server)
	if onServer {
		for _, device := range deviceIDList {
			if device.Name() != clientDeviceID {
				fileToClose, err := os.OpenFile(global.ConfigDir+global.PathSeparator+"deletions"+global.PathSeparator+device.Name()+global.FSSpace+strings.ReplaceAll(targetLocationIncomplete, "/", global.FSPath), os.O_CREATE|os.O_WRONLY, 0600)
				if err != nil {
					// do not print error as there is currently no way of seeing server-side errors
					// failure to add the target to the deletions list will exit the program and result in a client re-uploading the target (non-critical)
					os.Exit(back.ErrorWrite)
				}
				_ = fileToClose.Close() // error ignored; if the file could be created, it can probably be closed
			}
		}
	}

	// get the full targetLocation path and remove the target
	targetLocationComplete := global.TargetLocationFormat(targetLocationIncomplete)
	var isFile bool
	if !onServer { // error if target does not exist on client, needed because os.RemoveAll does not return an error if target does not exist
		isFile, _ = back.TargetIsFile(targetLocationComplete, true, 0)
	}
	err = os.RemoveAll(targetLocationComplete)
	if err != nil {
		return "", false, errors.New("unable to remove local target: " + err.Error())
	}

	if !onServer && len(deviceIDList) > 0 { // return the device ID if running on the client and a device ID exists (online mode)
		return (deviceIDList)[0].Name(), !isFile, nil
	}
	return "", true, nil

	// do not exit program, as this function is used as part of ShearRemoteFromClient
}

// RenameLocal renames oldLocationIncomplete to newLocationIncomplete on the local system.
// This function should only be used directly by the server binary.
func RenameLocal(oldLocationIncomplete, newLocationIncomplete string, verifyOldLocationExists bool) error {
	// get full paths for both locations
	oldLocation := global.TargetLocationFormat(oldLocationIncomplete)
	newLocation := global.TargetLocationFormat(newLocationIncomplete)

	if verifyOldLocationExists {
		back.TargetIsFile(oldLocation, true, 0)
	}

	// ensure newLocation does not exist
	_, isAccessible := back.TargetIsFile(newLocation, false, 0)
	if isAccessible {
		return errors.New("target already exists: " + newLocation)
	}

	// rename oldLocation to newLocation
	err := os.Rename(oldLocation, newLocation)
	if err != nil {
		return errors.New("unable to rename: " + err.Error())
	}

	return nil

	// do not exit program, as this function is used as part of RenameRemoteFromClient
}

// AddFolderLocal creates a new entry-containing directory on the local system.
// This function should only be used directly by the server binary.
func AddFolderLocal(targetLocationIncomplete string) error {
	// get the full targetLocation path and create the target
	targetLocationComplete := global.TargetLocationFormat(targetLocationIncomplete)
	err := os.Mkdir(targetLocationComplete, 0700)
	if err != nil {
		if os.IsExist(err) {
			fmt.Println(AnsiUpload + "Directory already exists - libmutton will still ensure it exists on the server")
		} else {
			return errors.New("unable to create directory: " + err.Error())
		}
	}

	return nil

	// do not exit program, as this function is used as part of AddFolderRemoteFromClient
}
