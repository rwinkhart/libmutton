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
