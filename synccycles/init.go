package synccycles

import (
	"errors"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rwinkhart/go-boilerplate/back"
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
	newDeviceIDPath := global.ConfigDir + global.PathSeparator + "devices" + global.PathSeparator + newDeviceID
	oldDeviceIDPath := global.ConfigDir + global.PathSeparator + "devices" + global.PathSeparator + oldDeviceID
	f, err := os.OpenFile(newDeviceIDPath, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return "", "", errors.New("unable to create local device ID file: " + err.Error())
	}
	_ = f.Close() // error ignored; if the file could be created, it can probably be closed

	cleanupOnFail := func() {
		// remove new device ID file
		os.RemoveAll(newDeviceIDPath)
		if oldDeviceID != global.FSMisc {
			// restore old device ID file (if it existed and has already been removed due to DirInit)
			if isAccessible, _ := back.TargetIsFile(oldDeviceIDPath, true); !isAccessible {
				f, _ := os.OpenFile(oldDeviceIDPath, os.O_CREATE|os.O_WRONLY, 0600)
				_ = f.Close() // error ignored; if the file could be created, it can probably be closed
			}
		}
	}

	// register new device ID with server and fetch remote EntryRoot and OS type
	// also removes the old device ID file (remotely)
	// if registration fails, remove the new device ID file locally and return before removing the old one
	sshClient, _, _, _, err := syncclient.GetSSHClient()
	if err != nil {
		cleanupOnFail()
		return "", "", errors.New("unable to connect to SSH client: " + err.Error())
	}
	output, err := syncclient.GetSSHOutput(sshClient, "libmuttonserver register", newDeviceID+"\n"+oldDeviceID)
	if err != nil {
		cleanupOnFail()
		return "", "", errors.New("unable to register device ID with server: " + err.Error())
	}
	sshEntryRootSSHIsWindows := strings.Split(output, global.FSSpace)
	_ = sshClient.Close() // ignore error; non-critical/unlikely/not much could be done about it

	// remove old device ID file (locally; may not exist)
	err = os.RemoveAll(oldDeviceIDPath)
	if err != nil {
		return "", "", errors.New("unable to remove old device ID file (locally): " + err.Error())
	}

	return sshEntryRootSSHIsWindows[0], sshEntryRootSSHIsWindows[1], nil
}
