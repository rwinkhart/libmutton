package syncclient

import (
	"errors"
	"strings"

	"github.com/rwinkhart/go-boilerplate/back"
	"github.com/rwinkhart/libmutton/global"
	"github.com/rwinkhart/libmutton/synccommon"
)

// ShearRemoteFromClient removes the target file or directory from
// the local system and calls the server to remove it remotely and
// add it to the deletions list.
// It can safely be called in offline mode, as well, so this is
// the intended interface for shearing (ShearLocal should only
// be used directly by the server binary).
func ShearRemoteFromClient(targetLocationIncomplete string) error {
	deviceID, isDir, err := synccommon.ShearLocal(targetLocationIncomplete, "") // remove the target from the local system and get the device ID of the client
	if err != nil {
		return errors.New("unable to shear target locally: " + err.Error())
	}

	sshClient, offlineMode, _, _, err := GetSSHClient()
	if offlineMode {
		goto end
	}
	if err != nil {
		return errors.New("unable to connect to SSH client: " + err.Error())
	}
	if deviceID == "" {
		return errors.New("unable to shear target remotely: no device ID found")
	}

	// ensure targetLocationIncomplete ends with a slash if it is a directory (for clarity in shear message)
	if isDir && !strings.HasSuffix(targetLocationIncomplete, "/") {
		targetLocationIncomplete += "/"
	}

	// call the server to remotely shear the target and add it to the deletions list
	_, err = GetSSHOutput(sshClient, "libmuttonserver shear", deviceID+"\n"+strings.ReplaceAll(targetLocationIncomplete, global.PathSeparator, global.FSPath))
	if err != nil {
		return errors.New("unable to shear target remotely: " + err.Error())
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

// RenameRemoteFromClient renames oldLocationIncomplete to newLocationIncomplete on
// the local system and calls the server to perform the rename remotely and add the
// old target to the deletions list.
// It can safely be called in offline mode, as well, so this is the intended
// interface for renaming (RenameLocal should only be used directly by the server binary).
func RenameRemoteFromClient(oldLocationIncomplete, newLocationIncomplete string) error {
	err := synccommon.RenameLocal(oldLocationIncomplete, newLocationIncomplete) // move the target on the local system
	if err != nil {
		return errors.New("unable to rename target locally: " + err.Error())
	}

	deviceIDList, err := global.GenDeviceIDList()
	if err != nil {
		return errors.New("unable to generate device ID list: " + err.Error())
	}
	// create an SSH client
	sshClient, offlineMode, _, _, err := GetSSHClient()
	if offlineMode {
		goto end
	}
	if err != nil {
		return errors.New("unable to connect to SSH client: " + err.Error())
	}
	if deviceIDList[0].Name() == "" {
		return errors.New("unable to rename target remotely: no device ID found")
	}

	// call the server to move the target on the remote system and add the old target to the deletions list
	_, err = GetSSHOutput(sshClient, "libmuttonserver rename",
		(deviceIDList)[0].Name()+"\n"+
			strings.ReplaceAll(oldLocationIncomplete, global.PathSeparator, global.FSPath)+"\n"+
			strings.ReplaceAll(newLocationIncomplete, global.PathSeparator, global.FSPath))
	if err != nil {
		return errors.New("unable to rename target remotely: " + err.Error())
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
func AddFolderRemoteFromClient(targetLocationIncomplete string) error {
	err := synccommon.AddFolderLocal(targetLocationIncomplete) // add the folder on the local system
	if err != nil {
		return errors.New("unable to add folder locally: " + err.Error())
	}

	// create an SSH client
	sshClient, offlineMode, _, _, err := GetSSHClient()
	if offlineMode {
		goto end
	}
	if err != nil {
		return errors.New("unable to connect to SSH client: " + err.Error())
	}

	// call the server to create the folder remotely
	_, err = GetSSHOutput(sshClient, "libmuttonserver addfolder", strings.ReplaceAll(targetLocationIncomplete, global.PathSeparator, global.FSPath)) // call the server to create the folder remotely
	if err != nil {
		return errors.New("unable to add folder remotely: " + err.Error())
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
