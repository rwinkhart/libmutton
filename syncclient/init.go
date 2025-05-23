package syncclient

import (
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rwinkhart/go-boilerplate/back"
	"github.com/rwinkhart/libmutton/core"
	"github.com/rwinkhart/libmutton/global"
)

// DeviceIDGen generates a new client device ID and registers it with the server (will replace existing one).
// Device IDs are only needed for online synchronization.
// Device IDs are guaranteed unique as the current UNIX time is appended to them.
// Returns: the remote EntryRoot and OS type indicator.
func DeviceIDGen(oldDeviceID string) (string, string) {
	// generate new device ID
	deviceIDPrefix, _ := os.Hostname()
	deviceIDSuffix := core.StringGen(rand.Intn(32)+48, 0.2, 1) + "-" + strconv.FormatInt(time.Now().Unix(), 10)
	newDeviceID := deviceIDPrefix + "-" + deviceIDSuffix

	// create new device ID file (locally)
	fileToClose, err := os.OpenFile(global.ConfigDir+global.PathSeparator+"devices"+global.PathSeparator+newDeviceID, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		back.PrintError("Failed to create local device ID file: "+err.Error(), back.ErrorWrite, true)
	}
	_ = fileToClose.Close() // error ignored; if the file could be created, it can probably be closed

	// remove old device ID file (locally; may not exist)
	err = os.RemoveAll(global.ConfigDir + global.PathSeparator + "devices" + global.PathSeparator + oldDeviceID)
	if err != nil {
		back.PrintError("Failed to remove old device ID file (locally): "+err.Error(), back.ErrorWrite, true)
	}

	// register new device ID with server and fetch remote EntryRoot and OS type
	// also removes the old device ID file (remotely)
	// manualSync is true so the user is alerted if device ID registration fails
	sshClient, _, _ := GetSSHClient(true)
	sshEntryRootSSHIsWindows := strings.Split(GetSSHOutput(sshClient, "libmuttonserver register", newDeviceID+"\n"+oldDeviceID), global.FSSpace)
	err = sshClient.Close()
	if err != nil {
		back.PrintError("Init failed - Unable to close SSH client: "+err.Error(), global.ErrorServerConnection, true)
	}

	return sshEntryRootSSHIsWindows[0], sshEntryRootSSHIsWindows[1]
}
