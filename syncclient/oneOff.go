package syncclient

import (
	"encoding/json"
	"errors"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rwinkhart/go-boilerplate/back"
	"github.com/rwinkhart/go-boilerplate/stringy"
	"github.com/rwinkhart/libmutton/global"
	"github.com/rwinkhart/libmutton/synccommon"
)

// ShearRemoteFromClient removes the target file or directory from
// the local system and calls the server to remove it remotely and
// add it to the deletions list.
// It can safely be called in offline mode, as well, so this is
// the intended interface for shearing (ShearLocal should only
// be used directly by the server binary).
func ShearRemoteFromClient(vanityPath string, onlyShearAgeFile bool) error {
	deviceID, isDir, err := synccommon.ShearLocal(vanityPath, "", onlyShearAgeFile) // remove the target from the local system and get the device ID of the client
	if err != nil {
		return errors.New("unable to shear target locally: " + err.Error())
	}

	var modifier string
	var output []byte
	sshClient, offlineMode, _, _, _, err := GetSSHClient()
	if offlineMode {
		goto end
	}
	if err != nil {
		return errors.New("unable to connect to SSH client: " + err.Error())
	}
	if deviceID == "" {
		return errors.New("unable to shear target remotely: no device ID found")
	}

	// ensure vanityPath ends with a slash if it is a directory (for clarity in shear message)
	if isDir && !strings.HasSuffix(vanityPath, "/") {
		vanityPath += "/"
	}

	// call the server to remotely shear the target and add it to the deletions list
	if onlyShearAgeFile {
		modifier = "-age"
	}
	output, err = GetSSHOutput(sshClient, "libmuttonserver shear"+modifier, deviceID+"\n"+strings.ReplaceAll(vanityPath, global.PathSeparator, global.FSPath))
	if err != nil {
		return errors.New("unable to shear target remotely: " + err.Error())
	}
	if len(output) > 11 {
		return errors.New("unable to complete shear; server-side error occurred: " + strings.ReplaceAll(string(output)[11:len(output)-2], global.FSSpace, "\n"))
	}

	// close the SSH client
	err = sshClient.Close()
	if err != nil {
		return errors.New("unable to close SSH client: " + err.Error())
	}

end:
	back.Exit(0) // sync is not required after shearing since the target has already been removed from the local system
	return nil
}

// RenameRemoteFromClient renames oldVanityPath to newVanityPath on
// the local system and calls the server to perform the rename remotely and add the
// old target to the deletions list.
// It can safely be called in offline mode, as well, so this is the intended
// interface for renaming (RenameLocal should only be used directly by the server binary).
func RenameRemoteFromClient(oldVanityPath, newVanityPath string) error {
	err := synccommon.RenameLocal(oldVanityPath, newVanityPath) // move the target on the local system
	if err != nil {
		return errors.New("unable to rename target locally: " + err.Error())
	}

	deviceIDList, err := global.GenDeviceIDList()
	if err != nil {
		return errors.New("unable to generate device ID list: " + err.Error())
	}
	if deviceIDList[0].Name() == "" {
		return errors.New("unable to rename target remotely: no device ID found")
	}

	// create an SSH client
	var output []byte
	sshClient, offlineMode, _, _, _, err := GetSSHClient()
	if offlineMode {
		goto end
	}
	if err != nil {
		return errors.New("unable to connect to SSH client: " + err.Error())
	}

	// call the server to move the target on the remote system and add the old target to the deletions list
	output, err = GetSSHOutput(sshClient, "libmuttonserver rename",
		(deviceIDList)[0].Name()+"\n"+
			strings.ReplaceAll(oldVanityPath, global.PathSeparator, global.FSPath)+"\n"+
			strings.ReplaceAll(newVanityPath, global.PathSeparator, global.FSPath))
	if err != nil {
		return errors.New("unable to rename target remotely: " + err.Error())
	}
	if len(output) > 11 {
		return errors.New("unable to complete rename; server-side error occurred: " + strings.ReplaceAll(string(output)[11:len(output)-2], global.FSSpace, "\n"))
	}

	// close the SSH client
	err = sshClient.Close()
	if err != nil {
		return errors.New("unable to close SSH client: " + err.Error())
	}

end:
	back.Exit(0)
	return nil
}

// AddFolderRemoteFromClient creates a new entry-containing directory
// on the local system and calls the server to create the folder remotely.
// It can safely be called in offline mode, as well, so this is the
// intended interface for adding folders (AddFolderLocal should only be
// used directly by the server binary).
func AddFolderRemoteFromClient(vanityPath string) error {
	err := synccommon.AddFolderLocal(vanityPath) // add the folder on the local system
	if err != nil {
		return errors.New("unable to add folder locally: " + err.Error())
	}

	// create an SSH client
	var output []byte
	sshClient, offlineMode, _, _, _, err := GetSSHClient()
	if offlineMode {
		goto end
	}
	if err != nil {
		return errors.New("unable to connect to SSH client: " + err.Error())
	}

	// call the server to create the folder remotely
	output, err = GetSSHOutput(sshClient, "libmuttonserver addfolder", strings.ReplaceAll(vanityPath, global.PathSeparator, global.FSPath)) // call the server to create the folder remotely
	if err != nil {
		return errors.New("unable to add folder remotely: " + err.Error())
	}
	if len(output) > 11 {
		return errors.New("unable to complete addfolder; server-side error occurred: " + strings.ReplaceAll(string(output)[11:len(output)-2], global.FSSpace, "\n"))
	}

	// close the SSH client
	err = sshClient.Close()
	if err != nil {
		return errors.New("unable to close SSH client: " + err.Error())
	}

end:
	back.Exit(0)
	return nil
}

// DeviceIDGenFromClient generates a new client device ID
// and registers it with the server (will replace existing one).
// Device IDs are only needed for online synchronization.
// Device IDs are guaranteed unique as the current UNIX time is appended to them.
// Leave prefix empty to use the current hostname as the prefix.
// Returns: the remote EntryRoot, the remote AgeDir, and OS type indicator.
func DeviceIDGenFromClient(oldDeviceID, prefix string) (string, string, string, error) {
	// generate new device ID
	if prefix == "" {
		prefix, _ = os.Hostname()
	}
	newDeviceID := prefix + "-" + stringy.StringGen(rand.Intn(32)+48, 0.2, 1) + "-" + strconv.FormatInt(time.Now().Unix(), 10)

	// create new device ID file (locally)
	newDeviceIDPath := global.ConfigDir + global.PathSeparator + "devices" + global.PathSeparator + newDeviceID
	oldDeviceIDPath := global.ConfigDir + global.PathSeparator + "devices" + global.PathSeparator + oldDeviceID
	f, err := os.OpenFile(newDeviceIDPath, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return "", "", "", errors.New("unable to create local device ID file: " + err.Error())
	}
	_ = f.Close() // error ignored; if the file could be created, it can probably be closed

	cleanupOnFail := func() {
		// remove new device ID file
		_ = os.RemoveAll(newDeviceIDPath)
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
	sshClient, _, _, _, _, err := GetSSHClient()
	if err != nil {
		cleanupOnFail()
		return "", "", "", errors.New("unable to connect to SSH client: " + err.Error())
	}
	output, err := GetSSHOutput(sshClient, "libmuttonserver register", newDeviceID+"\n"+oldDeviceID)
	if err != nil {
		cleanupOnFail()
		return "", "", "", errors.New("unable to register device ID with server: " + err.Error())
	}
	var registerResp synccommon.RegisterResp
	err = json.Unmarshal(output, &registerResp)
	if err != nil {
		cleanupOnFail()
		return "", "", "", errors.New("unable to unmarshal server register response: " + err.Error())
	}
	if registerResp.ErrMsg != nil {
		cleanupOnFail()
		return "", "", "", errors.New("unable to complete register; server-side error occurred: " + strings.ReplaceAll(*registerResp.ErrMsg, global.FSSpace, "\n"))
	}
	_ = sshClient.Close() // ignore error; non-critical/unlikely/not much could be done about it

	// remove old device ID file (locally; may not exist)
	err = os.RemoveAll(oldDeviceIDPath)
	if err != nil {
		cleanupOnFail()
		return "", "", "", errors.New("unable to remove old device ID file (locally): " + err.Error())
	}

	return registerResp.EntryRoot, registerResp.AgeDir, strconv.FormatBool(registerResp.IsWindows), nil
}
