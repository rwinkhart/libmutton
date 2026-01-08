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
	ErrMsg     *string    `json:"errMsg"` // nil if no error occurred
	ServerTime int64      `json:"serverTime"`
	Deletions  []Deletion `json:"deletions"`
	Entries    EntriesMap `json:"entries"`
}
type Deletion struct {
	VanityPath string `json:"vanityPath"`
	IsAgeFile  bool   `json:"isAgeFile"`
}
type EntriesMap map[string]Entry // map vanity paths to containing folders and mod+age timestamps
type Entry struct {
	ContainingFolder string `json:"containingFolder"`
	ModTime          int64  `json:"modTime"`
	AgeTimestamp     *int64 `json:"ageTimestamp"` // nil if no age file is present (non-password entry)
}

// RegisterResp defines the structure of responses from `libmuttonserver register`
type RegisterResp struct {
	ErrMsg    *string `json:"errMsg"` // nil if no error occurred
	EntryRoot string  `json:"entryRoot"`
	IsWindows bool    `json:"isWindows"`
}

// GetAllEntryData returns a map of all vanity paths to
// their respective containing folders and mod+age timestamps.
func GetAllEntryData() (EntriesMap, error) {
	var err error
	entryList, _, err := WalkEntryDir()
	if err != nil {
		return nil, errors.New("unable to walk entry directory: " + err.Error())
	}
	// initialize vanityPath keys in map
	outputEntries := make(EntriesMap)
	var modInfo, ageInfo os.FileInfo
	for _, vanityPath := range entryList {
		containingFolder := vanityPath[:strings.LastIndex(vanityPath, "/")]
		modInfo, err = os.Stat(global.GetRealPath(vanityPath))
		if err != nil {
			return nil, errors.New("unable to read mod time for " + vanityPath + ": " + err.Error())
		}
		ageInfo, err = os.Stat(global.GetRealAgePath(vanityPath))
		var ageTimestamp *int64
		if err == nil {
			ageTime := ageInfo.ModTime().Unix()
			ageTimestamp = &ageTime
		} else if !os.IsNotExist(err) {
			return nil, errors.New("unable to read age time for " + vanityPath + ": " + err.Error())
		}
		outputEntries[vanityPath] = Entry{ContainingFolder: containingFolder, ModTime: modInfo.ModTime().Unix(), AgeTimestamp: ageTimestamp}
	}
	return outputEntries, nil
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
					f, err := os.OpenFile(global.CfgDir+global.PathSeparator+"deletions"+global.PathSeparator+device.Name()+global.FSSpace+"entry"+global.FSSpace+strings.ReplaceAll(vanityPath, "/", global.FSPath), os.O_CREATE|os.O_WRONLY, 0600)
					if err != nil {
						// failure to add the target to the deletions list will exit the program and result in a client re-uploading the target (non-critical)
						return "", false, err
					}
					_ = f.Close() // error ignored; if the file could be created, it can probably be closed
				}
				f, err := os.OpenFile(global.CfgDir+global.PathSeparator+"deletions"+global.PathSeparator+device.Name()+global.FSSpace+"age"+global.FSSpace+strings.ReplaceAll(vanityPath, "/", global.FSPath), os.O_CREATE|os.O_WRONLY, 0600)
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
		if err = os.RemoveAll(realPath); err != nil {
			return "", false, errors.New("unable to remove local entry (" + vanityPath + "): " + err.Error())
		}
	}
	if err = ShearAgeFileLocal(vanityPath); err != nil {
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
	if err := os.RemoveAll(global.AgeDir + global.PathSeparator + strings.ReplaceAll(vanityPath, "/", global.FSPath)); err != nil {
		return errors.New("unable to remove age file for " + vanityPath + ": " + err.Error())
	}
	return nil
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
	if err := os.Rename(oldRealPath, newRealPath); err != nil {
		return errors.New("unable to rename: " + err.Error())
	}

	// do the same for the age file (if one exists) - also back up timestamp first
	fileInfo, err := os.Stat(oldRealAgePath)
	if err == nil { // assume age file does not exist if os.Stat errors
		if err = os.Rename(oldRealAgePath, newRealAgePath); err != nil {
			return errors.New("unable to rename: " + err.Error())
		}
		if err = os.Chtimes(newRealAgePath, time.Now(), fileInfo.ModTime()); err != nil {
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
	if err := os.Mkdir(realPath, 0700); err != nil {
		if os.IsExist(err) {
			fmt.Println(back.AnsiBlue + "Directory already exists - libmutton will still ensure it exists on the server")
		} else {
			return errors.New("unable to create directory: " + err.Error())
		}
	}

	return nil

	// do not exit program, as this function is used as part of AddFolderRemoteFromClient
}
