package sync

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rwinkhart/libmutton/core"
)

// DeviceIDGen generates a new client device ID and registers it with the server.
// Device IDs are only needed for online synchronization.
// Device IDs are guaranteed unique as the current UNIX time is appended to them.
// Returns: the remote EntryRoot and OS type indicator.
func DeviceIDGen() (string, string) {
	deviceIDPrefix, _ := os.Hostname()
	deviceIDSuffix := core.StringGen(rand.Intn(32)+48, true, 0.2, true) + "-" + strconv.FormatInt(time.Now().Unix(), 10)
	deviceID := deviceIDPrefix + "-" + deviceIDSuffix
	_, err := os.Create(core.ConfigDir + core.PathSeparator + "devices" + core.PathSeparator + deviceID) // TODO remove existing device ID file if it exists (from both client and server)
	if err != nil {
		fmt.Println(core.AnsiError + "Failed to create local device ID file: " + err.Error() + core.AnsiReset)
	}

	// register device ID with server and fetch remote EntryRoot and OS type
	//manualSync is true so the user is alerted if device ID registration fails
	sshClient, _, _ := GetSSHClient(true)
	sshEntryRootSSHIsWindows := strings.Split(GetSSHOutput(sshClient, "libmuttonserver register", deviceID), FSSpace)

	return sshEntryRootSSHIsWindows[0], sshEntryRootSSHIsWindows[1]
}
