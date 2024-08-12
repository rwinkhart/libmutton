package sync

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rwinkhart/libmutton/core"
	"golang.org/x/crypto/ssh"
)

// DeviceIDGen generates a new client device ID and registers it with the server (will replace existing one).
// Device IDs are only needed for online synchronization.
// Device IDs are guaranteed unique as the current UNIX time is appended to them.
// Returns: the remote EntryRoot and OS type indicator.
func DeviceIDGen(oldDeviceID string) (string, string) {
	// generate new device ID
	deviceIDPrefix, _ := os.Hostname()
	deviceIDSuffix := core.StringGen(rand.Intn(32)+48, true, 0.2, true) + "-" + strconv.FormatInt(time.Now().Unix(), 10)
	newDeviceID := deviceIDPrefix + "-" + deviceIDSuffix

	// create new device ID file (locally)
	_, err := os.Create(core.ConfigDir + core.PathSeparator + "devices" + core.PathSeparator + newDeviceID) // TODO remove existing device ID file if it exists (from both client and server)
	if err != nil {
		fmt.Println(core.AnsiError + "Failed to create local device ID file: " + err.Error() + core.AnsiReset)
		os.Exit(102)
	}

	// register new device ID with server and fetch remote EntryRoot and OS type
	// also removes the old device ID file (remotely)
	// manualSync is true so the user is alerted if device ID registration fails
	sshClient, _, _ := GetSSHClient(true)
	defer func(sshClient *ssh.Client) {
		err = sshClient.Close()
		if err != nil {
			fmt.Println(core.AnsiError + "Init failed - Unable to close SSH client: " + err.Error() + core.AnsiReset)
			os.Exit(104)
		}
	}(sshClient)
	sshEntryRootSSHIsWindows := strings.Split(GetSSHOutput(sshClient, "libmuttonserver register", newDeviceID+"\n"+oldDeviceID), FSSpace)

	return sshEntryRootSSHIsWindows[0], sshEntryRootSSHIsWindows[1]
}
