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
// Output is meant to be captured over SSH for interpretation by the client.
func GetRemoteDataFromServer(clientDeviceID string) {
	// collect info
	entryMap, err := synccommon.GetAllEntryData()
	if err != nil {
		fmt.Printf("{\"errMsg\":\"%s\"}", err.Error())
		return
	}
	deletionsList, err := os.ReadDir(global.CfgDir + global.PathSeparator + "deletions")
	if err != nil {
		fmt.Printf("{\"errMsg\":\"%s\"}", err.Error())
		return
	}

	// form response
	var fetchResp synccommon.FetchRespT
	//// server time
	fetchResp.ServerTime = time.Now().Unix()
	//// deletions
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
			if err = os.RemoveAll(global.CfgDir + global.PathSeparator + "deletions" + global.PathSeparator + deletion.Name()); err != nil {
				fmt.Printf("{\"errMsg\":\"%s\"}", err.Error())
				return
			}
		}
	}
	//// entries
	fetchResp.Entries = entryMap

	// marshal and print response to send to client
	fetchRespBytes, err := json.Marshal(fetchResp)
	if err != nil {
		fmt.Printf("{\"errMsg\":\"%s\"}", err.Error())
		return
	}
	fmt.Print(string(fetchRespBytes))
}
