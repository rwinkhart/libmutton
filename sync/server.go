package sync

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rwinkhart/libmutton/core"
)

// GetRemoteDataFromServer prints to stdout the remote entries, mod times, folders, and deletions.
// Lists in output are separated by FSSpace.
// Output is meant to be captured over SSH for interpretation by the client.
func GetRemoteDataFromServer(clientDeviceID string) {
	entryList, dirList := WalkEntryDir()
	modList := getModTimes(entryList)
	deletionsList, err := os.ReadDir(core.ConfigDir + core.PathSeparator + "deletions")
	if err != nil {
		core.PrintError("Failed to read the deletions directory: "+err.Error(), core.ErrorRead, true)
	}

	// print the current UNIX timestamp to stdout
	fmt.Print(time.Now().Unix())

	// print the lists to stdout
	// entry list
	fmt.Print(core.FSSpace)
	for _, entry := range entryList {
		fmt.Print(core.FSMisc + entry)
	}

	// modification time list
	fmt.Print(core.FSSpace)
	for _, mod := range modList {
		fmt.Print(core.FSMisc)
		fmt.Print(mod)
	}

	// directory/folder list
	fmt.Print(core.FSSpace)
	for _, dir := range dirList {
		fmt.Print(core.FSMisc + dir)
	}

	// deletions list
	fmt.Print(core.FSSpace)
	for _, deletion := range deletionsList {
		// print deletion if it is relevant to the current client device
		affectedIDTargetLocationIncomplete := strings.Split(deletion.Name(), core.FSSpace)
		if affectedIDTargetLocationIncomplete[0] == clientDeviceID {
			fmt.Print(core.FSMisc + strings.ReplaceAll(affectedIDTargetLocationIncomplete[1], core.FSPath, "/"))

			// assume successful client deletion and remove deletions file (if assumption is somehow false, worst case scenario is that the client will re-upload the deleted entry)
			_ = os.Remove(core.ConfigDir + core.PathSeparator + "deletions" + core.PathSeparator + deletion.Name()) // error ignored; function not run from a user-facing argument and thus the error would not be visible
		}
	}
}
