package sync

import (
	"strings"

	"github.com/rwinkhart/go-boilerplate/back"
	"github.com/rwinkhart/libmutton/core"
)

// ShearRemoteFromClient removes the target file or directory from the local system and calls the server to remove it remotely and add it to the deletions list.
// It can safely be called in offline mode, as well, so this is the intended interface for shearing (ShearLocal should only be used directly by the server binary).
func ShearRemoteFromClient(targetLocationIncomplete string, forceOffline bool) {
	deviceID, isDir := ShearLocal(targetLocationIncomplete, "") // remove the target from the local system and get the device ID of the client

	if !forceOffline && deviceID != "" { // ensure a device ID exists (online mode)
		// create an SSH client; manualSync is false in case a device ID exists but SSH is not configured
		sshClient, _, _ := GetSSHClient(false)

		// ensure targetLocationIncomplete ends with a slash if it is a directory (for clarity in shear message)
		if isDir && !strings.HasSuffix(targetLocationIncomplete, "/") {
			targetLocationIncomplete += "/"
		}

		// call the server to remotely shear the target and add it to the deletions list
		GetSSHOutput(sshClient, "libmuttonserver shear", deviceID+"\n"+strings.ReplaceAll(targetLocationIncomplete, core.PathSeparator, core.FSPath))

		// close the SSH client
		err := sshClient.Close()
		if err != nil {
			back.PrintError("Sync failed - Unable to close SSH client: "+err.Error(), core.ErrorServerConnection, true)
		}
	}

	back.Exit(0) // sync is not required after shearing since the target has already been removed from the local system
}

// RenameRemoteFromClient renames oldLocationIncomplete to newLocationIncomplete on the local system and calls the server to perform the rename remotely and add the old target to the deletions list.
// It can safely be called in offline mode, as well, so this is the intended interface for renaming (RenameLocal should only be used directly by the server binary).
func RenameRemoteFromClient(oldLocationIncomplete, newLocationIncomplete string, forceOffline bool) {
	RenameLocal(oldLocationIncomplete, newLocationIncomplete, false) // move the target on the local system

	deviceIDList := core.GenDeviceIDList(true)
	if !forceOffline && len(*deviceIDList) > 0 { // ensure a device ID exists (online mode)
		// create an SSH client; manualSync is false in case a device ID exists but SSH is not configured
		sshClient, _, _ := GetSSHClient(false)

		// call the server to move the target on the remote system and add the old target to the deletions list
		GetSSHOutput(sshClient, "libmuttonserver rename",
			(*deviceIDList)[0].Name()+"\n"+
				strings.ReplaceAll(oldLocationIncomplete, core.PathSeparator, core.FSPath)+"\n"+
				strings.ReplaceAll(newLocationIncomplete, core.PathSeparator, core.FSPath))

		// close the SSH client
		err := sshClient.Close()
		if err != nil {
			back.PrintError("Sync failed - Unable to close SSH client: "+err.Error(), core.ErrorServerConnection, true)
		}
	}

	back.Exit(0)
}

// AddFolderRemoteFromClient creates a new entry-containing directory on the local system and calls the server to create the folder remotely.
// It can safely be called in offline mode, as well, so this is the intended interface for adding folders (AddFolderLocal should only be used directly by the server binary).
func AddFolderRemoteFromClient(targetLocationIncomplete string, forceOffline bool) {
	AddFolderLocal(targetLocationIncomplete) // add the folder on the local system

	deviceIDList := core.GenDeviceIDList(true)
	if !forceOffline && len(*deviceIDList) > 0 { // ensure a device ID exists (online mode)
		// create an SSH client; manualSync is false in case a device ID exists but SSH is not configured
		sshClient, _, _ := GetSSHClient(false)

		// call the server to create the folder remotely
		GetSSHOutput(sshClient, "libmuttonserver addfolder", strings.ReplaceAll(targetLocationIncomplete, core.PathSeparator, core.FSPath)) // call the server to create the folder remotely

		// close the SSH client
		err := sshClient.Close()
		if err != nil {
			back.PrintError("Sync failed - Unable to close SSH client: "+err.Error(), core.ErrorServerConnection, true)
		}
	}

	back.Exit(0)
}
