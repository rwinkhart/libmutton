package sync

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"github.com/rwinkhart/libmutton/core"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// ANSI color constants used only in this file
const (
	ansiDelete   = "\033[38;5;1m"
	ansiDownload = "\033[38;5;2m"
	ansiUpload   = "\033[38;5;4m"
)

// GetSSHClient returns an SSH client connection to the server (also returns the remote EntryRoot and an indicator of the server's OS).
// Only supports key-based authentication (passphrases are supported for CLI-based implementations).
func GetSSHClient(manualSync bool) (*ssh.Client, string, bool) {
	// get SSH config info, exit if not configured (displaying an error if the sync job was called manually)
	var sshUserConfig []string
	var missingValueError string
	if manualSync {
		missingValueError = joinErrorWithEXE("SSH settings not configured - Run \"", " init\" to configure")
	} else {
		missingValueError = "0"
	}
	sshUserConfig = core.ParseConfig([][2]string{{"LIBMUTTON", "sshUser"}, {"LIBMUTTON", "sshIP"}, {"LIBMUTTON", "sshPort"}, {"LIBMUTTON", "sshKey"}, {"LIBMUTTON", "sshKeyProtected"}, {"LIBMUTTON", "sshEntryRoot"}, {"LIBMUTTON", "sshIsWindows"}}, missingValueError)

	var user, ip, port, keyFile, keyFileProtected, entryRoot string
	var isWindows bool
	var err error
	for i, key := range sshUserConfig {
		switch i {
		case 0:
			user = key
		case 1:
			ip = key
		case 2:
			port = key
		case 3:
			keyFile = key
		case 4:
			keyFileProtected = key
		case 5:
			entryRoot = key
		case 6:
			isWindows, err = strconv.ParseBool(key)
			if err != nil {
				fmt.Println(core.AnsiError + "Sync failed - Unable to parse server OS type: " + err.Error() + core.AnsiReset)
				os.Exit(101)
			}
		}
	}

	// read private key
	key, err := os.ReadFile(keyFile)
	if err != nil {
		fmt.Println(core.AnsiError+"Sync failed - Unable to read private key file:", keyFile+core.AnsiReset)
		os.Exit(101)
	}

	// parse private key
	var parsedKey ssh.Signer
	if keyFileProtected != "true" {
		parsedKey, err = ssh.ParsePrivateKey(key)
	} else {
		parsedKey, err = ssh.ParsePrivateKeyWithPassphrase(key, inputKeyFilePassphrase())
	}
	if err != nil {
		fmt.Println(core.AnsiError+"Sync failed - Unable to parse private key:", keyFile+core.AnsiReset)
		os.Exit(101)
	}

	// read known hosts file
	var hostKeyCallback ssh.HostKeyCallback
	hostKeyCallback, err = knownhosts.New(core.Home + core.PathSeparator + ".ssh" + core.PathSeparator + "known_hosts")
	if err != nil {
		fmt.Println(core.AnsiError + "Sync failed - Unable to read known hosts file: " + err.Error() + core.AnsiReset)
		os.Exit(101)
	}

	// configure SSH client
	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(parsedKey),
		},
		HostKeyCallback: hostKeyCallback,
	}

	// connect to SSH server
	sshClient, err := ssh.Dial("tcp", ip+":"+port, sshConfig)
	if err != nil {
		fmt.Println(core.AnsiError+"Sync failed - Unable to connect to remote server:", err.Error()+core.AnsiReset)
		os.Exit(104)
	}

	return sshClient, entryRoot, isWindows
}

// GetSSHOutput runs a command over SSH and returns the output as a string.
func GetSSHOutput(sshClient *ssh.Client, cmd, stdin string) string {
	// create a session
	sshSession, err := sshClient.NewSession()
	if err != nil {
		fmt.Println(core.AnsiError+"Sync failed - Unable to establish SSH session:", err.Error()+core.AnsiReset)
		os.Exit(104)
	}

	// provide stdin data for session
	sshSession.Stdin = strings.NewReader(stdin)

	// run the provided command
	var output []byte
	output, err = sshSession.CombinedOutput(cmd)
	if err != nil {
		fmt.Println(core.AnsiError+"Sync failed - Unable to run SSH command:", err.Error()+core.AnsiReset)
		os.Exit(103)
	}

	// convert the output to a string and remove leading/trailing whitespace
	outputString := string(output)
	outputString = strings.TrimSpace(outputString)

	return outputString
}

// getRemoteDataFromClient returns a map of remote entries to their modification times, a list of remote folders, a list of queued deletions, and the current server&client times as UNIX timestamps.
func getRemoteDataFromClient(sshClient *ssh.Client, manualSync bool) (map[string]int64, []string, []string, int64, int64) {
	// get remote output over SSH
	deviceIDList := core.GenDeviceIDList(true)
	if len(*deviceIDList) == 0 {
		if manualSync {
			fmt.Println(joinErrorWithEXE("Sync failed - No device ID found; run \"", " init\" to generate a device ID"))
			os.Exit(105)
		} else {
			core.Exit(0) // exit silently if the sync job was called automatically, as the user may just be in offline mode
		}
	}
	clientTime := time.Now().Unix() // get client time now to avoid accuracy issues caused by unpredictable sync time
	output := GetSSHOutput(sshClient, "libmuttonserver fetch", (*deviceIDList)[0].Name())

	// split output into slice based on occurrences of FSSpace
	outputSlice := strings.Split(output, FSSpace)

	// parse output/re-form lists
	if len(outputSlice) != 5 { // ensure information from server is complete
		fmt.Println(core.AnsiError + "Sync failed - Unable to fetch remote data; server returned an unexpected response" + core.AnsiReset)
		os.Exit(103)
	}
	serverTime, err := strconv.ParseInt(outputSlice[0], 10, 64)
	if err != nil {
		fmt.Println(core.AnsiError+"Sync failed - Unable to parse server time: ", err.Error()+core.AnsiReset)
		os.Exit(101)
	}
	entries := strings.Split(outputSlice[1], FSMisc)[1:]
	modsStrings := strings.Split(outputSlice[2], FSMisc)[1:]
	folders := strings.Split(outputSlice[3], FSMisc)[1:]
	deletions := strings.Split(outputSlice[4], FSMisc)[1:]

	// convert the mod times to int64
	var mods []int64
	var mod int64
	for _, modString := range modsStrings {
		mod, err = strconv.ParseInt(modString, 10, 64)
		if err != nil {
			fmt.Println(core.AnsiError+"Sync failed - Unable to parse mod time: ", err.Error()+core.AnsiReset)
			os.Exit(101)
		}
		mods = append(mods, mod)
	}

	// map remote entries to their modification times
	entryModMap := make(map[string]int64)
	for i, entry := range entries {
		entryModMap[entry] = mods[i]
	}

	return entryModMap, folders, deletions, serverTime, clientTime
}

// getLocalData returns a map of local entries to their modification times.
func getLocalData() map[string]int64 {
	// get a list of all entries
	entries, _ := WalkEntryDir()

	// get a list of all entry modification times
	modList := getModTimes(entries)

	// map the entries to their modification times
	entryModMap := make(map[string]int64)
	for i, entry := range entries {
		entryModMap[entry] = modList[i]
	}

	// return the lists
	return entryModMap
}

// targetLocationFormatSFTP formats the target location to match the remote server's entry directory and path separator.
func targetLocationFormatSFTP(targetName, serverEntryRoot string, serverIsWindows bool) string {
	if !serverIsWindows {
		return serverEntryRoot + targetName
	} else {
		return serverEntryRoot + strings.ReplaceAll(targetName, "/", "\\")
	}
}

// sftpSync takes two slices of entries (one for downloads and one for uploads) and syncs them between the client and server using SFTP.
func sftpSync(sshClient *ssh.Client, sshEntryRoot string, sshIsWindows bool, downloadList, uploadList []string) {
	// create an SFTP client from sshClient
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		fmt.Println(core.AnsiError+"Sync failed - Unable to establish SFTP session:", err.Error()+core.AnsiReset)
		os.Exit(104)
	}
	defer func(sftpClient *sftp.Client) {
		err = sftpClient.Close()
		if err != nil {
			fmt.Println(core.AnsiError+"Sync failed - Unable to close SFTP client: ", err.Error()+core.AnsiReset)
			os.Exit(104)
		}
	}(sftpClient)

	// iterate over the download list
	var filesTransferred bool
	for _, entryName := range downloadList {
		filesTransferred = true // set a flag to indicate that files have been downloaded (used to determine whether to print a gap between download and upload messages)

		fmt.Println("Downloading " + ansiDownload + entryName + core.AnsiReset)

		// store path to remote entry
		remoteEntryFullPath := targetLocationFormatSFTP(entryName, sshEntryRoot, sshIsWindows)

		// save modification time of remote file
		var fileInfo os.FileInfo
		fileInfo, err = sftpClient.Stat(remoteEntryFullPath)
		if err != nil {
			fmt.Println(core.AnsiError+"Sync failed - Unable to get remote file info (mod time):", err.Error()+core.AnsiReset)
			os.Exit(101)
		}
		modTime := fileInfo.ModTime()

		// open remote file
		var remoteFile *sftp.File
		remoteFile, err = sftpClient.Open(remoteEntryFullPath)
		if err != nil {
			fmt.Println(core.AnsiError+"Sync failed - Unable to open remote file:", err.Error()+core.AnsiReset)
			os.Exit(101)
		}

		// store path to local entry
		localEntryFullPath := core.TargetLocationFormat(entryName)

		// create local file
		var localFile *os.File
		localFile, err = os.Create(localEntryFullPath)
		if err != nil {
			fmt.Println(core.AnsiError+"Sync failed - Unable to create local file:", err.Error()+core.AnsiReset)
			os.Exit(102)
		}

		// download the file
		_, err = remoteFile.WriteTo(localFile)
		if err != nil {
			fmt.Println(core.AnsiError+"Sync failed - Unable to download remote file:", err.Error()+core.AnsiReset)
			os.Exit(103)
		}

		// close the files
		_ = remoteFile.Close() // errors ignored; if the files could be opened/created, it can probably be closed
		_ = localFile.Close()

		// set the modification time of the local file to match the value saved from the remote file (from before the download)
		err = os.Chtimes(localEntryFullPath, time.Now(), modTime)
	}

	if filesTransferred {
		fmt.Println() // add a gap between download and upload messages
	}

	// iterate over the upload list
	filesTransferred = false
	for _, entryName := range uploadList {
		filesTransferred = true // set a flag to indicate that files have been uploaded (used to determine whether to print a gap between upload and sync complete messages)

		fmt.Println("Uploading " + ansiUpload + entryName + core.AnsiReset)

		// store path to local entry
		localEntryFullPath := core.TargetLocationFormat(entryName)

		// save modification time of local file
		var fileInfo os.FileInfo
		fileInfo, err = os.Stat(localEntryFullPath)
		if err != nil {
			fmt.Println(core.AnsiError+"Sync failed - Unable to get local file info (mod time):", err.Error()+core.AnsiReset)
			os.Exit(101)
		}
		modTime := fileInfo.ModTime()

		// open local file
		var localFile *os.File
		localFile, err = os.Open(localEntryFullPath)
		if err != nil {
			fmt.Println(core.AnsiError+"Sync failed - Unable to open local file:", err.Error()+core.AnsiReset)
			os.Exit(101)
		}

		// store path to remote entry
		remoteEntryFullPath := targetLocationFormatSFTP(entryName, sshEntryRoot, sshIsWindows)

		// create remote file
		var remoteFile *sftp.File
		remoteFile, err = sftpClient.Create(remoteEntryFullPath)
		if err != nil {
			fmt.Println(core.AnsiError+"Sync failed - Unable to create remote file ("+remoteEntryFullPath+"):", err.Error()+core.AnsiReset)
			os.Exit(102)
		}

		// upload the file
		_, err = localFile.WriteTo(remoteFile)
		if err != nil {
			fmt.Println(core.AnsiError+"Sync failed - Unable to upload local file:", err.Error()+core.AnsiReset)
			os.Exit(103)
		}

		// close the files
		_ = localFile.Close() // errors ignored; if the files could be opened/created, it can probably be closed
		_ = remoteFile.Close()

		// set the modification time of the remote file to match the value saved from the local file (from before the upload)
		err = sftpClient.Chtimes(remoteEntryFullPath, time.Now(), modTime)
	}

	if filesTransferred {
		fmt.Println() // add a gap between upload and sync complete messages
	}
}

// syncLists determines which entries need to be downloaded and uploaded for synchronizations and calls sftpSync with this information.
// Using maps means that syncing will be done in an arbitrary order, but it is a worthy tradeoff for speed and simplicity.
func syncLists(sshClient *ssh.Client, sshEntryRoot string, sshIsWindows, timeSynced bool, localEntryModMap, remoteEntryModMap map[string]int64) {
	// initialize slices to store entries that need to be downloaded or uploaded
	var downloadList, uploadList []string

	// iterate over client entries
	for entry, localModTime := range localEntryModMap {
		// check if the entry is present in the server map
		if remoteModTime, present := remoteEntryModMap[entry]; present {
			// entry exists on both client and server, compare mod times
			if remoteModTime > localModTime {
				fmt.Println(ansiDownload+entry+core.AnsiReset, "is newer on server, adding to download list")
				downloadList = append(downloadList, entry)
			} else if remoteModTime < localModTime {
				fmt.Println(ansiUpload+entry+core.AnsiReset, "is newer on client, adding to upload list")
				uploadList = append(uploadList, entry)
			}
			// remove entry from remoteEntryModMap (process of elimination)
			delete(remoteEntryModMap, entry)
		} else {
			fmt.Println(ansiUpload+entry+core.AnsiReset, "does not exist on server, adding to upload list")
			uploadList = append(uploadList, entry)
		}
	}

	// iterate over remaining entries in remoteEntryModMap
	for entry := range remoteEntryModMap {
		fmt.Println(ansiDownload+entry+core.AnsiReset, "does not exist on client, adding to download list")
		downloadList = append(downloadList, entry)
	}

	// call sftpSync with the download and upload lists
	if timeSynced && (max(len(downloadList), len(uploadList)) > 0) { // only call sftpSync if there are entries to download or upload
		fmt.Println() // add a gap between list-add messages and the actual sync messages from sftpSync
		sftpSync(sshClient, sshEntryRoot, sshIsWindows, downloadList, uploadList)
	} else if !timeSynced {
		// do not call sftpSync if the client and server times are out of sync
		core.Exit(1)
	}

	fmt.Println("Client is synchronized with server")
}

// deletionSync removes entries from the client that have been deleted on the server (multi-client deletion).
func deletionSync(deletions []string) {
	var filesDeleted bool
	for _, deletion := range deletions {
		filesDeleted = true // set a flag to indicate that files have been deleted (used to determine whether to print a gap between deletion and other messages)
		fmt.Println(ansiDelete+deletion+core.AnsiReset, "has been sheared, removing locally (if it exists)")
		err := os.RemoveAll(core.TargetLocationFormat(deletion))
		if err != nil {
			fmt.Println(core.AnsiError + "Sync failed - Failed to shear " + deletion + " locally: " + err.Error() + core.AnsiReset)
			os.Exit(102)
		}
	}

	if filesDeleted {
		fmt.Println() // add a gap between deletion and other messages
	}
}

// ShearRemoteFromClient removes the target file or directory from the local system and calls the server to remove it remotely and add it to the deletions list.
// It can safely be called in offline mode, as well, so this is the intended interface for shearing (ShearLocal should only be used directly by the server binary).
func ShearRemoteFromClient(sshClient *ssh.Client, targetLocationIncomplete string) {
	deviceID := ShearLocal(targetLocationIncomplete, "") // remove the target from the local system and get the device ID of the client

	if deviceID != "" { // ensure a device ID exists (online mode)
		// call the server to remotely shear the target and add it to the deletions list
		GetSSHOutput(sshClient, "libmuttonserver shear", deviceID+"\n"+strings.ReplaceAll(targetLocationIncomplete, core.PathSeparator, FSPath))
	}

	core.Exit(0) // sync is not required after shearing since the target has already been removed from the local system
}

// RenameRemoteFromClient renames oldLocationIncomplete to newLocationIncomplete on the local system and calls the server to perform the rename remotely and add the old target to the deletions list.
// It can safely be called in offline mode, as well, so this is the intended interface for renaming (RenameLocal should only be used directly by the server binary).
func RenameRemoteFromClient(sshClient *ssh.Client, oldLocationIncomplete, newLocationIncomplete string) {
	RenameLocal(oldLocationIncomplete, newLocationIncomplete, false) // move the target on the local system

	deviceIDList := core.GenDeviceIDList(true)
	if len(*deviceIDList) > 0 { // ensure a device ID exists (online mode)
		// call the server to move the target on the remote system and add the old target to the deletions list
		GetSSHOutput(sshClient, "libmuttonserver rename",
			(*deviceIDList)[0].Name()+"\n"+
				strings.ReplaceAll(oldLocationIncomplete, core.PathSeparator, FSPath)+"\n"+
				strings.ReplaceAll(newLocationIncomplete, core.PathSeparator, FSPath))
	}

	core.Exit(0)
}

// AddFolderRemoteFromClient creates a new entry-containing directory on the local system and calls the server to create the folder remotely.
// It can safely be called in offline mode, as well, so this is the intended interface for adding folders (AddFolderLocal should only be used directly by the server binary).
func AddFolderRemoteFromClient(sshClient *ssh.Client, targetLocationIncomplete string) {
	AddFolderLocal(targetLocationIncomplete)                                                                                       // add the folder on the local system
	GetSSHOutput(sshClient, "libmuttonserver addfolder", strings.ReplaceAll(targetLocationIncomplete, core.PathSeparator, FSPath)) // call the server to create the folder remotely

	core.Exit(0)
}

// folderSync creates folders on the client (from the given list of folder names).
func folderSync(folders []string) {
	for _, folder := range folders {
		// store the full local path of the folder
		folderFullPath := core.TargetLocationFormat(folder)

		// check if folder already exists
		isFile, isAccessible := core.TargetIsFile(folderFullPath, false, 1)

		if !isFile && !isAccessible {
			err := os.MkdirAll(folderFullPath, 0700)
			if err != nil {
				fmt.Println(core.AnsiError + "Sync failed - Failed to create folder \"" + folder + "\": " + err.Error() + core.AnsiReset)
				os.Exit(102)
			}
		} else if isFile {
			fmt.Println(core.AnsiError + "Sync failed - Failed to create folder \"" + folder + "\" - A file with the same name already exists" + core.AnsiReset)
			os.Exit(106)
		}
	}
}

// RunJob runs the SSH sync job.
// Setting manualSync to true will throw errors if sync is not configured (online mode is assumed).
func RunJob(manualSync bool) {
	// get SSH client to re-use throughout the sync process
	sshClient, sshEntryRoot, sshIsWindows := GetSSHClient(manualSync)
	defer func(sshClient *ssh.Client) {
		err := sshClient.Close()
		if err != nil {
			fmt.Println(core.AnsiError+"Sync failed - Unable to close SSH client: ", err.Error()+core.AnsiReset)
			os.Exit(104)
		}
	}(sshClient)

	// fetch remote lists
	remoteEntryModMap, remoteFolders, deletions, serverTime, clientTime := getRemoteDataFromClient(sshClient, manualSync)

	// sync folders
	folderSync(remoteFolders)

	// sync deletions
	deletionSync(deletions)

	// fetch local lists
	localEntryModMap := getLocalData()

	// prior to syncing lists, ensure the client and server clocks are synced within 45 seconds
	var timeSynced = true
	timeDiff := serverTime - clientTime
	if timeDiff < -45 || timeDiff > 45 {
		timeSynced = false
		fmt.Print(core.AnsiError + "Client and server clocks are out of sync.\n\nPlease ensure both clocks are correct before attempting to sync again.\n\nA dry sync output will be printed below (if any operations would have been performed). It is strongly recommended to review it and manually update the modification times as applicable to ensure the correct version of each entry is kept.\n\nIf the client's clock is at fault, update the modification times of any entries pending upload, even if the correct (upload) operation is being performed on them. Failure to do so could result in entries being uploaded to the server with the incorrect modification times (could result in data loss).\n\n" + core.AnsiReset)
	}

	// sync new and updated entries
	syncLists(sshClient, sshEntryRoot, sshIsWindows, timeSynced, localEntryModMap, remoteEntryModMap)

	// exit program after successful sync
	core.Exit(0)
}
