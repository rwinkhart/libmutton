package sync

import (
	"fmt"
	"os"
	"strings"

	"github.com/rwinkhart/libmutton/core"
)

// ShearRemoteFromClient removes the target file or directory from the local system and calls the server to remove it remotely and add it to the deletions list.
// It can safely be called in offline mode, as well, so this is the intended interface for shearing (ShearLocal should only be used directly by the server binary).
func ShearRemoteFromClient(targetLocationIncomplete string) {
	deviceID := ShearLocal(targetLocationIncomplete, "") // remove the target from the local system and get the device ID of the client

	if deviceID != "" { // ensure a device ID exists (online mode)
		// create an SSH client; manualSync is false in case a device ID exists but SSH is not configured
		sshClient, _, _ := GetSSHClient(false)

		// call the server to remotely shear the target and add it to the deletions list
		GetSSHOutput(sshClient, "libmuttonserver shear", deviceID+"\n"+strings.ReplaceAll(targetLocationIncomplete, core.PathSeparator, FSPath))

		// close the SSH client
		err := sshClient.Close()
		if err != nil {
			fmt.Println(core.AnsiError+"Sync failed - Unable to close SSH client:", err.Error()+core.AnsiReset)
			os.Exit(104)
		}
	}

	core.Exit(0) // sync is not required after shearing since the target has already been removed from the local system
}

// RenameRemoteFromClient renames oldLocationIncomplete to newLocationIncomplete on the local system and calls the server to perform the rename remotely and add the old target to the deletions list.
// It can safely be called in offline mode, as well, so this is the intended interface for renaming (RenameLocal should only be used directly by the server binary).
func RenameRemoteFromClient(oldLocationIncomplete, newLocationIncomplete string) {
	RenameLocal(oldLocationIncomplete, newLocationIncomplete, false) // move the target on the local system

	deviceIDList := core.GenDeviceIDList(true)
	if len(*deviceIDList) > 0 { // ensure a device ID exists (online mode)
		// create an SSH client; manualSync is false in case a device ID exists but SSH is not configured
		sshClient, _, _ := GetSSHClient(false)

		// call the server to move the target on the remote system and add the old target to the deletions list
		GetSSHOutput(sshClient, "libmuttonserver rename",
			(*deviceIDList)[0].Name()+"\n"+
				strings.ReplaceAll(oldLocationIncomplete, core.PathSeparator, FSPath)+"\n"+
				strings.ReplaceAll(newLocationIncomplete, core.PathSeparator, FSPath))

		// close the SSH client
		err := sshClient.Close()
		if err != nil {
			fmt.Println(core.AnsiError+"Sync failed - Unable to close SSH client:", err.Error()+core.AnsiReset)
			os.Exit(104)
		}
	}

	core.Exit(0)
}

// AddFolderRemoteFromClient creates a new entry-containing directory on the local system and calls the server to create the folder remotely.
// It can safely be called in offline mode, as well, so this is the intended interface for adding folders (AddFolderLocal should only be used directly by the server binary).
func AddFolderRemoteFromClient(targetLocationIncomplete string) {
	AddFolderLocal(targetLocationIncomplete) // add the folder on the local system

	deviceIDList := core.GenDeviceIDList(true)
	if len(*deviceIDList) > 0 { // ensure a device ID exists (online mode)
		// create an SSH client; manualSync is false in case a device ID exists but SSH is not configured
		sshClient, _, _ := GetSSHClient(false)

		// call the server to create the folder remotely
		GetSSHOutput(sshClient, "libmuttonserver addfolder", strings.ReplaceAll(targetLocationIncomplete, core.PathSeparator, FSPath)) // call the server to create the folder remotely

		// close the SSH client
		err := sshClient.Close()
		if err != nil {
			fmt.Println(core.AnsiError+"Sync failed - Unable to close SSH client:", err.Error()+core.AnsiReset)
			os.Exit(104)
		}
	}

	core.Exit(0)
}
