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
		fmt.Println(core.AnsiError+"Failed to read the deletions directory:", err.Error()+core.AnsiReset)
		os.Exit(101)
	}

	// print the current UNIX timestamp to stdout
	fmt.Print(time.Now().Unix())

	// print the lists to stdout
	// entry list
	fmt.Print(FSSpace)
	for _, entry := range entryList {
		fmt.Print(FSMisc + entry)
	}

	// modification time list
	fmt.Print(FSSpace)
	for _, mod := range modList {
		fmt.Print(FSMisc)
		fmt.Print(mod)
	}

	// directory/folder list
	fmt.Print(FSSpace)
	for _, dir := range dirList {
		fmt.Print(FSMisc + dir)
	}

	// deletions list
	fmt.Print(FSSpace)
	for _, deletion := range deletionsList {
		// print deletion if it is relevant to the current client device
		affectedIDTargetLocationIncomplete := strings.Split(deletion.Name(), FSSpace)
		if affectedIDTargetLocationIncomplete[0] == clientDeviceID {
			fmt.Print(FSMisc + strings.ReplaceAll(affectedIDTargetLocationIncomplete[1], FSPath, "/"))

			// assume successful client deletion and remove deletions file (if assumption is somehow false, worst case scenario is that the client will re-upload the deleted entry)
			_ = os.Remove(core.ConfigDir + core.PathSeparator + "deletions" + core.PathSeparator + deletion.Name()) // error ignored; function not run from a user-facing argument and thus the error would not be visible
		}
	}
}
