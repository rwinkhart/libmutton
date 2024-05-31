package sync

import (
	"fmt"
	"github.com/rwinkhart/MUTN/src/backend"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// WalkEntryDir walks the entry directory and returns lists of all files and directories found (two separate lists)
func WalkEntryDir() ([]string, []string) {
	// define file/directory containing slices so that they may be accessed by the anonymous WalkDir function
	var fileList []string
	var dirList []string

	// walk entry directory
	_ = filepath.WalkDir(backend.EntryRoot,
		func(fullPath string, entry fs.DirEntry, err error) error {

			// check for errors encountered while walking directory
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Println(backend.AnsiError+"The entry directory does not exist - run \""+os.Args[0], "init"+"\" to create it"+backend.AnsiReset) // TODO implement init command for libmuttonserver
				} else {
					// otherwise, print the source of the error
					fmt.Println(backend.AnsiError + "An unexpected error occurred while generating the entry list: " + err.Error() + backend.AnsiReset)
				}
				os.Exit(1)
			}

			// trim root path from each path before storing
			trimmedPath := fullPath[rootLength:]

			// create separate slices for entries and directories
			if !entry.IsDir() {
				fileList = append(fileList, trimmedPath)
			} else {
				dirList = append(dirList, trimmedPath)
			}

			return nil
		})

	return fileList, dirList
}

func getModTimes(entryList []string) []int64 {
	// get a list of all entry modification times
	var modList []int64
	for _, file := range entryList {
		modTime, _ := os.Stat(backend.TargetLocationFormat(file))
		modList = append(modList, modTime.ModTime().Unix())
	}

	return modList
}

// ShearLocal removes the target file or directory from the local system
// returns: deviceID (on client), for use in ShearRemoteFromClient
// if the local system is a server, it will also add the target to the deletions list for all clients (except the requesting client)
// this function should only be used directly by the server binary
func ShearLocal(targetLocationIncomplete, clientDeviceID string) string {
	// determine if running on a server
	var onServer bool
	if clientDeviceID != "" {
		onServer = true
	}

	// create a slice of all registered devices
	deviceIDList, err := os.ReadDir(backend.ConfigDir + backend.PathSeparator + "devices")
	if err != nil {
		fmt.Println(backend.AnsiError + "Failed to read the devices directory: " + err.Error() + backend.AnsiReset)
		os.Exit(1)
	}

	// add the sheared target (incomplete, vanity) to the deletions list (if running on a server)
	if onServer {
		for _, device := range deviceIDList {
			if device.Name() != clientDeviceID {
				_, err = os.Create(backend.ConfigDir + backend.PathSeparator + "deletions" + backend.PathSeparator + device.Name() + "\x1d" + strings.ReplaceAll(targetLocationIncomplete, backend.PathSeparator, "\x1e"))
				if err != nil {
					// do not print error as there is currently no way of seeing server-side errors
					// failure to add the target to the deletions list will exit the program and result in a client re-uploading the target (non-critical)
					os.Exit(1)
				}
			}
		}
	}

	// get the full targetLocation path and remove the target
	targetLocationComplete := backend.TargetLocationFormat(targetLocationIncomplete)
	if !onServer { // error if target does not exist on client, needed because os.RemoveAll does not return an error if target does not exist
		backend.TargetIsFile(targetLocationComplete, true, 0)
	}
	err = os.RemoveAll(targetLocationComplete)
	if err != nil {
		fmt.Println(backend.AnsiError + "Failed to remove local target: " + err.Error() + backend.AnsiReset)
		os.Exit(1)
	}

	if !onServer && len(deviceIDList) > 0 { // return the device ID if running on the client and a device ID exists (online mode)
		return deviceIDList[0].Name()
	}
	return ""

	// do not exit program, as this function is used as part of ShearRemoteFromClient
}

// AddFolderLocal creates a new entry-containing directory on the local system
// this function should only be used directly by the server binary
func AddFolderLocal(targetLocationIncomplete string) {
	// get the full targetLocation path and create the target
	targetLocationComplete := backend.TargetLocationFormat(targetLocationIncomplete)
	err := os.Mkdir(targetLocationComplete, 0700)
	if err != nil {
		if os.IsExist(err) {
			fmt.Println(backend.AnsiError + "Directory already exists" + backend.AnsiReset)
			os.Exit(1)
		} else {
			fmt.Println(backend.AnsiError + "Failed to create directory: " + err.Error() + backend.AnsiReset)
			os.Exit(1)
		}
	}

	// do not exit program, as this function is used as part of AddFolderRemoteFromClient
}
