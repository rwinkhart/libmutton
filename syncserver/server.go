package syncserver

import (
	"encoding/json"
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
	entryList, dirList, err := synccommon.WalkEntryDir()
	if err != nil {
		fmt.Printf("{\"errMsg\":\"%s\"}", err.Error())
		return
	}
	modList := synccommon.GetModTimes(entryList)
	deletionsList, err := os.ReadDir(global.ConfigDir + global.PathSeparator + "deletions")
	if err != nil {
		fmt.Printf("{\"errMsg\":\"%s\"}", err.Error())
		return
	}
	vanityPathsToTimestamps, err := synccommon.GetEntryAges()
	if err != nil {
		fmt.Printf("{\"errMsg\":\"%s\"}", err.Error())
		return
	}
	var fetchResp synccommon.FetchResp

	// server time
	fetchResp.ServerTime = time.Now().Unix()

	// folders (initialize keys in map)
	fetchResp.FoldersToEntries = make(map[string][]synccommon.Entry)
	for _, folder := range dirList {
		if _, exists := fetchResp.FoldersToEntries[folder]; !exists {
			fetchResp.FoldersToEntries[folder] = []synccommon.Entry{}
		}
	}

	// entries
	var folder string
	for i := range entryList {
		var ageTimestamp *int64
		if timestamp, exists := vanityPathsToTimestamps[entryList[i]]; exists {
			ageTimestamp = &timestamp
		}
		folder = entryList[i][:strings.LastIndex(entryList[i], "/")]
		fetchResp.FoldersToEntries[folder] = append(fetchResp.FoldersToEntries[folder], synccommon.Entry{VanityPath: entryList[i], ModTime: modList[i], AgeTimestamp: ageTimestamp})
	}

	// deletions
	for _, deletion := range deletionsList {
		// perform deletion if it is relevant to the current client device
		affectedIDVanityPath := strings.Split(deletion.Name(), global.FSSpace)
		if affectedIDVanityPath[0] == clientDeviceID {
			var isAgeFile bool
			if affectedIDVanityPath[1] == "age" {
				isAgeFile = true
			}
			fetchResp.Deletions = append(fetchResp.Deletions, synccommon.Deletion{VanityPath: strings.ReplaceAll(affectedIDVanityPath[2], global.FSPath, "/"), IsAgeFile: isAgeFile})

			// assume successful client deletion and remove deletions file (if assumption is somehow false, worst case scenario is that the client will re-upload the deleted entry)
			err = os.RemoveAll(global.ConfigDir + global.PathSeparator + "deletions" + global.PathSeparator + deletion.Name()) // error ignored; function not run from a user-facing argument and thus the error would not be visible
			if err != nil {
				fmt.Printf("{\"errMsg\":\"%s\"}", err.Error())
				return
			}
		}
	}

	// marshal and print response to send to client
	fetchRespBytes, err := json.Marshal(fetchResp)
	if err != nil {
		fmt.Printf("{\"errMsg\":\"%s\"}", err.Error())
		return
	}
	fmt.Print(string(fetchRespBytes))
}
