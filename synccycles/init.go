package synccycles

import (
	"errors"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rwinkhart/libmutton/global"
	"github.com/rwinkhart/libmutton/syncclient"
)

// DeviceIDGen generates a new client device ID and registers it with the server (will replace existing one).
// Device IDs are only needed for online synchronization.
// Device IDs are guaranteed unique as the current UNIX time is appended to them.
// Returns: the remote EntryRoot and OS type indicator.
func DeviceIDGen(oldDeviceID string) (string, string, error) {
	// generate new device ID
	deviceIDPrefix, _ := os.Hostname()
	deviceIDSuffix := StringGen(rand.Intn(32)+48, 0.2, 1) + "-" + strconv.FormatInt(time.Now().Unix(), 10)
	newDeviceID := deviceIDPrefix + "-" + deviceIDSuffix

	// create new device ID file (locally)
	fileToClose, err := os.OpenFile(global.ConfigDir+global.PathSeparator+"devices"+global.PathSeparator+newDeviceID, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return "", "", errors.New("failed to create local device ID file: " + err.Error())
	}
	_ = fileToClose.Close() // error ignored; if the file could be created, it can probably be closed

	// remove old device ID file (locally; may not exist)
	err = os.RemoveAll(global.ConfigDir + global.PathSeparator + "devices" + global.PathSeparator + oldDeviceID)
	if err != nil {
		return "", "", errors.New("failed to remove old device ID file (locally): " + err.Error())
	}

	// register new device ID with server and fetch remote EntryRoot and OS type
	// also removes the old device ID file (remotely)
	// manualSync is true so the user is alerted if device ID registration fails
	sshClient, _, _, err := syncclient.GetSSHClient(true)
	if err != nil {
		return "", "", errors.New("device ID gen failed - unable to connect to SSH client: " + err.Error())
	}
	sshEntryRootSSHIsWindows := strings.Split(syncclient.GetSSHOutput(sshClient, "libmuttonserver register", newDeviceID+"\n"+oldDeviceID), global.FSSpace)
	err = sshClient.Close()
	if err != nil {
		return "", "", errors.New("device ID gen failed - unable to close SSH client: " + err.Error())
	}

	return sshEntryRootSSHIsWindows[0], sshEntryRootSSHIsWindows[1], nil
}
