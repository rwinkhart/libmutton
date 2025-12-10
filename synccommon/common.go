package synccommon

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rwinkhart/go-boilerplate/back"
	"github.com/rwinkhart/libmutton/global"
)

// ANSI color constants used only in this file
const (
	AnsiDelete = "\033[38;5;1m"
)

// FetchResp defines the structure of responses from `libmuttonserver fetch`.
type FetchResp struct {
	ErrMsg           *string            `json:"errMsg"` // nil if no error occurred
	ServerTime       int64              `json:"serverTime"`
	Deletions        []Deletion         `json:"deletions"`
	FoldersToEntries map[string][]Entry `json:"folders"`
}
type Deletion struct {
	VanityPath string `json:"vanityPath"`
	IsAgeFile  bool   `json:"isAgeFile"`
}
type Entry struct {
	VanityPath   string `json:"vanityPath"`
	ModTime      int64  `json:"modTime"`
	AgeTimestamp *int64 `json:"ageTimestamp"` // nil if no age file is present (non-password entry)
}

// RegisterResp defines the structure of responses from `libmuttonserver register`
type RegisterResp struct {
	ErrMsg    *string `json:"errMsg"` // nil if no error occurred
	EntryRoot string  `json:"entryRoot"`
	AgeDir    string  `json:"ageDir"`
	IsWindows bool    `json:"isWindows"`
}

// GetModTimes returns a list of all entry modification times.
func GetModTimes(entryList []string) []int64 {
	var modList []int64
	for _, file := range entryList {
		modTime, _ := os.Stat(global.GetRealPath(file))
		modList = append(modList, modTime.ModTime().Unix())
	}

	return modList
}

// ShearLocal removes the target file or directory from the local system.
// Returns: deviceID (only on client; for use in ShearRemoteFromClient),
// isDir (only on client; for use in ShearRemoteFromClient).
// If the local system is a server, it will also add the target to the deletions list for all clients (except the requesting client).
// This function should only be used directly by the server binary.
func ShearLocal(vanityPath, clientDeviceID string, onlyShearAgeFile bool) (string, bool, error) {
	// determine if running on a server
	var onServer bool
	if clientDeviceID != "" {
		onServer = true
	}

	deviceIDList, err := global.GenDeviceIDList()
	if err != nil {
		return "", false, errors.New("unable to generate device ID list: " + err.Error())
	}

	// add the sheared vanityPath to the deletions list (if running on a server)
	if onServer {
		for _, device := range deviceIDList {
			if device.Name() != clientDeviceID {
				if !onlyShearAgeFile {
					f, err := os.OpenFile(global.ConfigDir+global.PathSeparator+"deletions"+global.PathSeparator+device.Name()+global.FSSpace+"entry"+global.FSSpace+strings.ReplaceAll(vanityPath, "/", global.FSPath), os.O_CREATE|os.O_WRONLY, 0600)
					if err != nil {
						// failure to add the target to the deletions list will exit the program and result in a client re-uploading the target (non-critical)
						return "", false, err
					}
					_ = f.Close() // error ignored; if the file could be created, it can probably be closed
				}
				f, err := os.OpenFile(global.ConfigDir+global.PathSeparator+"deletions"+global.PathSeparator+device.Name()+global.FSSpace+"age"+global.FSSpace+strings.ReplaceAll(vanityPath, "/", global.FSPath), os.O_CREATE|os.O_WRONLY, 0600)
				if err != nil {
					// failure to add the target to the deletions list will exit the program and result in a client re-uploading the target (non-critical)
					return "", false, err
				}
				_ = f.Close() // error ignored; if the file could be created, it can probably be closed
			}
		}
	}

	// remove the target locally
	realPath := global.GetRealPath(vanityPath)
	var isFile bool
	if !onServer { // error if target does not exist on client, needed because os.RemoveAll does not return an error if target does not exist
		isAccessible, err := back.TargetIsFile(realPath, true)
		if !isAccessible {
			return "", false, err
		}
		if err == nil { // fails if target is a directory, so no error indicates a file
			isFile = true
		}
	}
	if !onlyShearAgeFile {
		err = os.RemoveAll(realPath)
		if err != nil {
			return "", false, errors.New("unable to remove local entry (" + vanityPath + "): " + err.Error())
		}
	}
	err = ShearAgeFileLocal(vanityPath)
	if err != nil {
		return "", false, err
	}

	if !onServer && len(deviceIDList) > 0 { // return the device ID if running on the client and a device ID exists (online mode)
		return (deviceIDList)[0].Name(), !isFile, nil
	}
	return "", true, nil

	// do not exit program, as this function is used as part of ShearRemoteFromClient
}

// ShearAgeFileLocal removes the age file for a vanity path.
// This function should only be used directly by the server binary.
func ShearAgeFileLocal(vanityPath string) error {
	err := os.RemoveAll(global.AgeDir + global.PathSeparator + strings.ReplaceAll(vanityPath, "/", global.FSPath))
	if err != nil {
		return errors.New("unable to remove age file for " + vanityPath + ": " + err.Error())
	}
	return nil
}

// GetEntryAges reads the age directory and returns a
// map of vanity paths to their corresponding age timestamps.
func GetEntryAges() (map[string]int64, error) {
	contents, err := os.ReadDir(global.AgeDir)
	if err != nil {
		return nil, errors.New("unable to read age directory contents: " + err.Error())
	}

	var vanityPathsToTimestamps = make(map[string]int64)
	for _, dirEntry := range contents {
		if !dirEntry.IsDir() {
			vanityPath := strings.ReplaceAll(dirEntry.Name(), global.FSPath, "/")
			info, err := dirEntry.Info()
			if err != nil {
				return nil, errors.New("unable to read age file modtime for " + vanityPath + ": " + err.Error())
			}
			vanityPathsToTimestamps[vanityPath] = info.ModTime().Unix()
		}
	}

	return vanityPathsToTimestamps, nil
}

// RenameLocal renames oldLocationIncomplete to newLocationIncomplete on the local system.
// This function should only be used directly by the server binary.
func RenameLocal(oldVanityPath, newVanityPath string) error {
	// get full paths for both locations
	oldRealPath := global.GetRealPath(oldVanityPath)
	oldRealAgePath := global.GetRealAgePath(oldVanityPath)
	newRealPath := global.GetRealPath(newVanityPath)
	newRealAgePath := global.GetRealAgePath(newVanityPath)

	// ensure newLocation does not exist
	isAccessible, _ := back.TargetIsFile(newRealPath, true) // error is ignored because dir/file status is irrelevant
	if isAccessible {
		return errors.New("new target (" + newRealPath + ") already exists")
	}

	// rename oldLocation to newLocation
	err := os.Rename(oldRealPath, newRealPath)
	if err != nil {
		return errors.New("unable to rename: " + err.Error())
	}

	// do the same for the age file (if one exists) - also back up timestamp first
	var fileInfo os.FileInfo
	fileInfo, err = os.Stat(oldRealAgePath)
	if err == nil { // assume age file does not exist if os.Stat errors
		err = os.Rename(oldRealAgePath, newRealAgePath)
		if err != nil {
			return errors.New("unable to rename: " + err.Error())
		}
		err = os.Chtimes(newRealAgePath, time.Now(), fileInfo.ModTime())
		if err != nil {
			return errors.New("unable to set timestamp on age file for " + newVanityPath + ": " + err.Error())
		}
	}

	return nil
	// do not exit program, as this function is used as part of RenameRemoteFromClient
}

// AddFolderLocal creates a new entry-containing directory on the local system.
// This function should only be used directly by the server binary.
func AddFolderLocal(vanityPath string) error {
	// create the target locally
	realPath := global.GetRealPath(vanityPath)
	err := os.Mkdir(realPath, 0700)
	if err != nil {
		if os.IsExist(err) {
			fmt.Println(back.AnsiBlue + "Directory already exists - libmutton will still ensure it exists on the server")
		} else {
			return errors.New("unable to create directory: " + err.Error())
		}
	}

	return nil

	// do not exit program, as this function is used as part of AddFolderRemoteFromClient
}
