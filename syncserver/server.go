package syncserver

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rwinkhart/libmutton/global"
	"github.com/rwinkhart/libmutton/synccommon"
)

// GetRemoteDataFromServer prints to stdout the remote entries, mod times, folders, and deletions.
// Lists in output are separated by FSSpace.
// Output is meant to be captured over SSH for interpretation by the client.
func GetRemoteDataFromServer(clientDeviceID string) {
	entryList, dirList, _ := synccommon.WalkEntryDir()
	modList := synccommon.GetModTimes(entryList)
	deletionsList, _ := os.ReadDir(global.ConfigDir + global.PathSeparator + "deletions")
	vanityPathsToTimestamps, _ := synccommon.GetEntryAges()
	var ageVanityPaths []string
	var ageTimestamps []int64
	for vanityPath, timestamp := range vanityPathsToTimestamps {
		ageVanityPaths = append(ageVanityPaths, vanityPath)
		ageTimestamps = append(ageTimestamps, timestamp)
	}

	// print the current UNIX timestamp to stdout
	fmt.Print(time.Now().Unix())

	// print the lists to stdout
	// entry list
	fmt.Print(global.FSSpace)
	for _, entry := range entryList {
		fmt.Print(global.FSMisc + entry)
	}

	// modification time list
	fmt.Print(global.FSSpace)
	for _, mod := range modList {
		fmt.Print(global.FSMisc)
		fmt.Print(mod)
	}

	// age file list
	fmt.Print(global.FSSpace)
	for _, file := range ageVanityPaths {
		fmt.Print(global.FSMisc + file)
	}

	// age file timestamp list
	fmt.Print(global.FSSpace)
	for _, timestamp := range ageTimestamps {
		fmt.Print(global.FSMisc)
		fmt.Print(timestamp)
	}

	// directory/folder list
	fmt.Print(global.FSSpace)
	for _, dir := range dirList {
		fmt.Print(global.FSMisc + dir)
	}

	// deletions list
	fmt.Print(global.FSSpace)
	for _, deletion := range deletionsList {
		// perform deletion if it is relevant to the current client device
		affectedIDVanityPath := strings.Split(deletion.Name(), global.FSSpace)
		if affectedIDVanityPath[0] == clientDeviceID {
			fmt.Print(global.FSMisc + strings.ReplaceAll(affectedIDVanityPath[1]+global.FSSpace+affectedIDVanityPath[2], global.FSPath, "/"))

			// assume successful client deletion and remove deletions file (if assumption is somehow false, worst case scenario is that the client will re-upload the deleted entry)
			_ = os.RemoveAll(global.ConfigDir + global.PathSeparator + "deletions" + global.PathSeparator + deletion.Name()) // error ignored; function not run from a user-facing argument and thus the error would not be visible
		}
	}
}
