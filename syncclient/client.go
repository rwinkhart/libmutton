package syncclient

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"github.com/rwinkhart/go-boilerplate/back"
	"github.com/rwinkhart/libmutton/cfg"
	"github.com/rwinkhart/libmutton/global"
	"github.com/rwinkhart/libmutton/synccommon"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// GetSSHClient
// Returns:
// sshClient,
// offlineMode (whether the client is in offline mode).
// sshIsWindows (whether the remote server is running Windows),
// sshEntryRoot (the root directory for entries on the remote server),
// Only supports key-based authentication (passphrases are supported for CLI-based implementations).
func GetSSHClient() (*ssh.Client, bool, bool, string, error) {
	// get SSH config info
	sshUserConfig, err := cfg.ParseConfig([][2]string{{"LIBMUTTON", "offlineMode"}, {"LIBMUTTON", "sshUser"}, {"LIBMUTTON", "sshIP"}, {"LIBMUTTON", "sshPort"}, {"LIBMUTTON", "sshKey"}, {"LIBMUTTON", "sshKeyProtected"}, {"LIBMUTTON", "sshEntryRoot"}, {"LIBMUTTON", "sshIsWindows"}})
	if len(sshUserConfig) == 1 {
		// offline mode is enabled
		return nil, true, false, "", nil
	}
	if err != nil {
		return nil, false, false, "", errors.New("unable to parse SSH config: " + err.Error())
	}

	var user, ip, port, keyFile, keyFileProtected, entryRoot string
	var isWindows bool
	for i, key := range sshUserConfig {
		switch i {
		case 1:
			user = key
		case 2:
			ip = key
		case 3:
			port = key
		case 4:
			keyFile = key
		case 5:
			keyFileProtected = key
		case 6:
			entryRoot = key
		case 7:
			isWindows, err = strconv.ParseBool(key)
			if err != nil {
				return nil, false, false, "", errors.New("unable to parse server OS type: " + err.Error())
			}
		}
	}

	// read private key
	key, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, false, false, "", errors.New("unable to read private key: " + keyFile)
	}

	// parse private key
	var parsedKey ssh.Signer
	if keyFileProtected != "true" {
		parsedKey, err = ssh.ParsePrivateKey(key)
	} else {
		parsedKey, err = ssh.ParsePrivateKeyWithPassphrase(key, global.GetPassphrase("Enter passphrase for your SSH keyfile:"))
	}
	if err != nil {
		return nil, false, false, "", errors.New("unable to parse private key: " + keyFile)
	}

	// read known hosts file
	var hostKeyCallback ssh.HostKeyCallback
	hostKeyCallback, err = knownhosts.New(back.Home + global.PathSeparator + ".ssh" + global.PathSeparator + "known_hosts")
	if err != nil {
		return nil, false, false, "", errors.New("unable to read known hosts file: " + err.Error())
	}

	// configure SSH client
	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(parsedKey),
		},
		HostKeyCallback: hostKeyCallback,
		Timeout:         3 * time.Second,
	}

	// connect to SSH server
	sshClient, err := ssh.Dial("tcp", ip+":"+port, sshConfig)
	if err != nil {
		return nil, false, false, "", errors.New("unable to connect to remote server: " + err.Error())
	}

	return sshClient, false, isWindows, entryRoot, nil
}

// GetSSHOutput runs a command over SSH and returns the output as a string.
func GetSSHOutput(sshClient *ssh.Client, cmd, stdin string) (string, error) {
	// create a session
	sshSession, err := sshClient.NewSession()
	if err != nil {
		return "", errors.New("unable to establish SSH session: " + err.Error())
	}

	// provide stdin data for session
	sshSession.Stdin = strings.NewReader(stdin)

	// run the provided command
	var output []byte
	output, err = sshSession.CombinedOutput(cmd)
	if err != nil {
		return "", errors.New("unable to run SSH command: " + err.Error())
	}

	// convert the output to a string and remove leading/trailing whitespace
	outputString := string(output)
	outputString = strings.TrimSpace(outputString)

	return outputString, nil
}

// getRemoteDataFromClient returns a map of remote entries to their modification times, a list of remote folders, a list of queued deletions, and the current server&client times as UNIX timestamps.
func getRemoteDataFromClient(sshClient *ssh.Client) (map[string]int64, []string, []string, int64, int64, error) {
	// get remote output over SSH
	deviceIDList, err := global.GenDeviceIDList()
	if err != nil {
		return nil, nil, nil, 0, 0, err
	}
	if len(deviceIDList) == 0 {
		return nil, nil, nil, 0, 0, errors.New("no device ID found")
	}
	clientTime := time.Now().Unix() // get client time now to avoid accuracy issues caused by unpredictable sync time
	output, err := GetSSHOutput(sshClient, "libmuttonserver fetch", (deviceIDList)[0].Name())
	if err != nil {
		return nil, nil, nil, 0, 0, errors.New("unable to run remote command: " + err.Error())
	}

	// split output into slice based on occurrences of FSSpace
	outputSlice := strings.Split(output, global.FSSpace)

	// parse output/re-form lists
	if len(outputSlice) != 5 { // ensure information from server is complete
		return nil, nil, nil, 0, 0, errors.New("unable to run remote command; server returned an unexpected response")
	}
	serverTime, err := strconv.ParseInt(outputSlice[0], 10, 64)
	if err != nil {
		return nil, nil, nil, 0, 0, errors.New("unable to parse server time: " + err.Error())
	}
	entries := strings.Split(outputSlice[1], global.FSMisc)[1:]
	modsStrings := strings.Split(outputSlice[2], global.FSMisc)[1:]
	folders := strings.Split(outputSlice[3], global.FSMisc)[1:]
	deletions := strings.Split(outputSlice[4], global.FSMisc)[1:]

	// convert the mod times to int64
	var mods []int64
	var mod int64
	for _, modString := range modsStrings {
		mod, err = strconv.ParseInt(modString, 10, 64)
		if err != nil {
			return nil, nil, nil, 0, 0, errors.New("unable to parse mod time: " + err.Error())
		}
		mods = append(mods, mod)
	}

	// map remote entries to their modification times
	entryModMap := make(map[string]int64)
	for i, entry := range entries {
		entryModMap[entry] = mods[i]
	}

	return entryModMap, folders, deletions, serverTime, clientTime, nil
}

// getLocalData returns a map of local entries to their modification times.
func getLocalData() (map[string]int64, error) {
	// get a list of all entries
	entries, _, err := synccommon.WalkEntryDir()
	if err != nil {
		return nil, err
	}

	// get a list of all entry modification times
	modList := synccommon.GetModTimes(entries)

	// map the entries to their modification times
	entryModMap := make(map[string]int64)
	for i, entry := range entries {
		entryModMap[entry] = modList[i]
	}

	// return the lists
	return entryModMap, nil
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
func sftpSync(sshClient *ssh.Client, sshEntryRoot string, sshIsWindows bool, downloadList, uploadList []string) error {
	// create an SFTP client from sshClient
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return errors.New("unable to establish SFTP session: " + err.Error())
	}
	defer func(sftpClient *sftp.Client) {
		_ = sftpClient.Close()
	}(sftpClient)

	// iterate over the download list
	var filesTransferred bool
	for _, entryName := range downloadList {
		filesTransferred = true // set a flag to indicate that files have been downloaded (used to determine whether to print a gap between download and upload messages)

		fmt.Println("Downloading " + synccommon.AnsiDownload + entryName + back.AnsiReset)

		// store path to remote entry
		remoteEntryFullPath := targetLocationFormatSFTP(entryName, sshEntryRoot, sshIsWindows)

		// save modification time of remote file
		var fileInfo os.FileInfo
		fileInfo, err = sftpClient.Stat(remoteEntryFullPath)
		if err != nil {
			return errors.New("unable to get remote file info (mod time): " + err.Error())
		}
		modTime := fileInfo.ModTime()

		// open remote file
		var remoteFile *sftp.File
		remoteFile, err = sftpClient.Open(remoteEntryFullPath)
		if err != nil {
			return errors.New("unable to open remote file: " + err.Error())
		}

		// store path to local entry
		localEntryFullPath := global.TargetLocationFormat(entryName)

		// create local file
		var localFile *os.File
		localFile, err = os.OpenFile(localEntryFullPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
		if err != nil {
			return errors.New("unable to create local file: " + err.Error())
		}

		// download the file
		_, err = remoteFile.WriteTo(localFile)
		if err != nil {
			return errors.New("unable to download remote file: " + err.Error())
		}

		// close the files
		_ = remoteFile.Close() // errors ignored; if the files could be opened/created, it can probably be closed
		_ = localFile.Close()

		// set the modification time of the local file to match the value saved from the remote file (from before the download)
		err = os.Chtimes(localEntryFullPath, time.Now(), modTime)
		if err != nil {
			return errors.New("unable to set local file modification time: " + err.Error())
		}
	}

	if filesTransferred {
		fmt.Println() // add a gap between download and upload messages
	}

	// iterate over the upload list
	filesTransferred = false
	for _, entryName := range uploadList {
		filesTransferred = true // set a flag to indicate that files have been uploaded (used to determine whether to print a gap between upload and sync complete messages)

		fmt.Println("Uploading " + synccommon.AnsiUpload + entryName + back.AnsiReset)

		// store path to local entry
		localEntryFullPath := global.TargetLocationFormat(entryName)

		// save modification time of local file
		var fileInfo os.FileInfo
		fileInfo, err = os.Stat(localEntryFullPath)
		if err != nil {
			return errors.New("unable to get local file info (mod time): " + err.Error())
		}
		modTime := fileInfo.ModTime()

		// open local file
		var localFile *os.File
		localFile, err = os.Open(localEntryFullPath)
		if err != nil {
			return errors.New("unable to open local file: " + err.Error())
		}

		// store path to remote entry
		remoteEntryFullPath := targetLocationFormatSFTP(entryName, sshEntryRoot, sshIsWindows)

		// create remote file
		var remoteFile *sftp.File
		remoteFile, err = sftpClient.OpenFile(remoteEntryFullPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY)
		if err != nil {
			return errors.New("unable to create remote file: " + err.Error())
		}

		// upload the file
		_, err = localFile.WriteTo(remoteFile)
		if err != nil {
			return errors.New("unable to upload local file: " + err.Error())
		}

		// close the files
		_ = localFile.Close() // errors ignored; if the files could be opened/created, it can probably be closed
		_ = remoteFile.Close()

		// set permissions on remote file
		err = sftpClient.Chmod(remoteEntryFullPath, 0600)
		if err != nil {
			return errors.New("unable to set permissions on remote file: " + err.Error())
		}

		// set the modification time of the remote file to match the value saved from the local file (from before the upload)
		err = sftpClient.Chtimes(remoteEntryFullPath, time.Now(), modTime)
		if err != nil {
			return errors.New("unable to set remote file modification time: " + err.Error())
		}
	}

	if filesTransferred {
		fmt.Println() // add a gap between upload and sync complete messages
	}

	return nil
}

// syncLists determines which entries need to be downloaded and uploaded for synchronizations and calls sftpSync with this information.
// Using maps means that syncing will be done in an arbitrary order, but it is a worthy tradeoff for speed and simplicity.
func syncLists(sshClient *ssh.Client, sshEntryRoot string, sshIsWindows, timeSynced, returnLists bool, localEntryModMap, remoteEntryModMap map[string]int64) ([3][]string, error) {
	// initialize slices to store entries that need to be downloaded or uploaded
	var downloadList, uploadList []string

	// iterate over client entries
	for entry, localModTime := range localEntryModMap {
		// check if the entry is present in the server map
		if remoteModTime, present := remoteEntryModMap[entry]; present {
			// entry exists on both client and server, compare mod times
			if remoteModTime > localModTime {
				fmt.Println(synccommon.AnsiDownload+entry+back.AnsiReset, "is newer on server, adding to download list")
				downloadList = append(downloadList, entry)
			} else if remoteModTime < localModTime {
				fmt.Println(synccommon.AnsiUpload+entry+back.AnsiReset, "is newer on client, adding to upload list")
				uploadList = append(uploadList, entry)
			}
			// remove entry from remoteEntryModMap (process of elimination)
			delete(remoteEntryModMap, entry)
		} else {
			fmt.Println(synccommon.AnsiUpload+entry+back.AnsiReset, "does not exist on server, adding to upload list")
			uploadList = append(uploadList, entry)
		}
	}

	// iterate over remaining entries in remoteEntryModMap
	for entry := range remoteEntryModMap {
		fmt.Println(synccommon.AnsiDownload+entry+back.AnsiReset, "does not exist on client, adding to download list")
		downloadList = append(downloadList, entry)
	}

	// call sftpSync with the download and upload lists
	if timeSynced && (max(len(downloadList), len(uploadList)) > 0) { // only call sftpSync if there are entries to download or upload
		fmt.Println() // add a gap between list-add messages and the actual sync messages from sftpSync
		err := sftpSync(sshClient, sshEntryRoot, sshIsWindows, downloadList, uploadList)
		if err != nil {
			return [3][]string{nil, nil, nil}, errors.New("unable to sync entries: " + err.Error())
		}
	} else if !timeSynced {
		// do not call sftpSync if the client and server times are out of sync
		back.Exit(global.ErrorSyncProcess)
	}

	fmt.Println("Client is synchronized with server")

	if returnLists {
		return [3][]string{nil, downloadList, uploadList}, nil
	}
	return [3][]string{nil, nil, nil}, nil
}

// deletionSync removes entries from the client that have been deleted on the server (multi-client deletion).
func deletionSync(deletions []string) error {
	var filesDeleted bool
	for _, deletion := range deletions {
		filesDeleted = true // set a flag to indicate that files have been deleted (used to determine whether to print a gap between deletion and other messages)
		fmt.Println(synccommon.AnsiDelete+deletion+back.AnsiReset, "has been sheared, removing locally (if it exists)")
		err := os.RemoveAll(global.TargetLocationFormat(deletion))
		if err != nil {
			return errors.New("unable to shear " + deletion + " locally: " + err.Error())
		}
	}
	if filesDeleted {
		fmt.Println() // add a gap between deletion and other messages
	}
	return nil
}

// folderSync creates folders on the client (from the given list of folder names).
func folderSync(folders []string) error {
	for _, folder := range folders {
		// store the full local path of the folder
		folderFullPath := global.TargetLocationFormat(folder)

		// check if target path already exists
		isAccessible, err := back.TargetIsFile(folderFullPath, false)

		if !isAccessible {
			err := os.MkdirAll(folderFullPath, 0700)
			if err != nil {
				return errors.New("unable to create folder (" + folder + "): " + err.Error())
			}
		} else if err != nil {
			return errors.New("unable to create folder (" + folder + "): " + err.Error())
		}
	}
	return nil
}

// RunJob runs the SSH sync job.
// Setting returnLists to true will return the deletions, downloads, and uploads lists for use by the client.
func RunJob(returnLists bool) ([3][]string, error) {
	// get SSH client to re-use throughout the sync process
	sshClient, offlineMode, sshIsWindows, sshEntryRoot, err := GetSSHClient()
	if offlineMode {
		return [3][]string{nil, nil, nil}, nil
	}
	if err != nil {
		return [3][]string{nil, nil, nil}, errors.New("unable to connect to SSH client: " + err.Error())
	}
	defer func(sshClient *ssh.Client) {
		_ = sshClient.Close()
	}(sshClient)

	// fetch remote lists
	remoteEntryModMap, remoteFolders, deletions, serverTime, clientTime, err := getRemoteDataFromClient(sshClient)
	if err != nil {
		return [3][]string{nil, nil, nil}, errors.New("unable to fetch remote data: " + err.Error())
	}

	// sync deletions
	err = deletionSync(deletions)
	if err != nil {
		return [3][]string{nil, nil, nil}, errors.New("unable to sync deletions: " + err.Error())
	}

	// sync folders
	err = folderSync(remoteFolders)
	if err != nil {
		return [3][]string{nil, nil, nil}, errors.New("unable to sync folders: " + err.Error())
	}

	// fetch local lists
	localEntryModMap, err := getLocalData()
	if err != nil {
		return [3][]string{nil, nil, nil}, errors.New("unable to fetch local entry data: " + err.Error())
	}

	// before syncing lists, ensure the client and server clocks are synced within 45 seconds
	var timeSynced = true
	timeDiff := serverTime - clientTime
	if timeDiff < -45 || timeDiff > 45 {
		timeSynced = false
		fmt.Print(back.AnsiError + "Client and server clocks are out of sync.\n\nPlease ensure both clocks are correct before attempting to sync again.\n\nA dry sync output will be printed below (if any operations would have been performed). It is strongly recommended to review it and manually update the modification times as applicable to ensure the correct version of each entry is kept.\n\nIf the client's clock is at fault, update the modification times of any entries pending upload, even if the correct (upload) operation is being performed on them. Failure to do so could result in entries being uploaded to the server with the incorrect modification times (could result in data loss).\n\n" + back.AnsiReset)
	}

	// sync new and updated entries
	var lists [3][]string
	if returnLists {
		lists, err = syncLists(sshClient, sshEntryRoot, sshIsWindows, timeSynced, true, localEntryModMap, remoteEntryModMap)
		if err != nil {
			return [3][]string{nil, nil, nil}, errors.New("unable to sync entries: " + err.Error())
		}
		lists[0] = deletions
		return lists, nil
	}
	_, err = syncLists(sshClient, sshEntryRoot, sshIsWindows, timeSynced, false, localEntryModMap, remoteEntryModMap)
	if err != nil {
		return [3][]string{nil, nil, nil}, errors.New("unable to sync entries: " + err.Error())
	}
	back.Exit(0)      // exit program if running non-interactively
	return lists, nil // dummy return for when not returning lists
}
