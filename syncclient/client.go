package syncclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
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
// sshAgeDir (the directory housing age files on the remote server),
// Only supports key-based authentication (passwords are supported for CLI-based implementations).
func GetSSHClient() (*ssh.Client, bool, *bool, *string, *string, error) {
	// get SSH config info
	config, err := cfg.LoadConfig()
	if err != nil {
		return nil, false, nil, nil, nil, errors.New("unable to parse SSH config: " + err.Error())
	}
	if *config.Libmutton.OfflineMode {
		return nil, true, nil, nil, nil, nil
	}

	// read private key
	key, err := os.ReadFile(*config.Libmutton.SSHKeyPath)
	if err != nil {
		return nil, false, nil, nil, nil, errors.New("unable to read private key: " + *config.Libmutton.SSHKeyPath)
	}

	// parse private key
	var parsedKey ssh.Signer
	if !*config.Libmutton.SSHKeyProtected {
		parsedKey, err = ssh.ParsePrivateKey(key)
	} else {
		parsedKey, err = ssh.ParsePrivateKeyWithPassphrase(key, global.GetPassword("Enter password for your SSH keyfile:"))
	}
	if err != nil {
		return nil, false, nil, nil, nil, errors.New("unable to parse private key: " + *config.Libmutton.SSHKeyPath)
	}

	// read known hosts file
	var hostKeyCallback ssh.HostKeyCallback
	hostKeyCallback, err = knownhosts.New(back.Home + global.PathSeparator + ".ssh" + global.PathSeparator + "known_hosts")
	if err != nil {
		return nil, false, nil, nil, nil, errors.New("unable to read known hosts file: " + err.Error())
	}

	// configure SSH client
	sshConfig := &ssh.ClientConfig{
		User: *config.Libmutton.SSHUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(parsedKey),
		},
		HostKeyCallback: hostKeyCallback,
		Timeout:         3 * time.Second,
	}

	// connect to SSH server
	sshClient, err := ssh.Dial("tcp", *config.Libmutton.SSHIP+":"+*config.Libmutton.SSHPort, sshConfig)
	if err != nil {
		return nil, false, nil, nil, nil, errors.New("unable to connect to remote server: " + err.Error())
	}

	return sshClient, false, config.Libmutton.SSHIsWindows, config.Libmutton.SSHEntryRootPath, config.Libmutton.SSHAgeDirPath, nil
}

// GetSSHOutput runs a command over SSH and returns the output as a string.
func GetSSHOutput(sshClient *ssh.Client, cmd, stdin string) ([]byte, error) {
	// create a session
	sshSession, err := sshClient.NewSession()
	if err != nil {
		return nil, errors.New("unable to establish SSH session: " + err.Error())
	}

	// provide stdin data for session
	sshSession.Stdin = strings.NewReader(stdin)

	// run the provided command
	var output []byte
	output, err = sshSession.CombinedOutput(cmd)
	if err != nil {
		return nil, errors.New("unable to run SSH command: " + err.Error())
	}

	return output, nil
}

// getRemoteDataFromClient returns:
// a map of remote entries to their modification times,
// a map of remote entries to their timestamps,
// a list of remote folders,
// a list of queued deletions,
// and the current server&client times as UNIX timestamps.
func getRemoteDataFromClient(sshClient *ssh.Client) (map[string]int64, map[string]int64, []string, []synccommon.Deletion, int64, int64, error) {
	// get remote output over SSH
	deviceIDList, err := global.GenDeviceIDList()
	if err != nil {
		return nil, nil, nil, nil, 0, 0, err
	}
	if len(deviceIDList) == 0 {
		return nil, nil, nil, nil, 0, 0, errors.New("no device ID found")
	}
	clientTime := time.Now().Unix() // get client time now to avoid accuracy issues caused by unpredictable sync time
	output, err := GetSSHOutput(sshClient, "libmuttonserver fetch", (deviceIDList)[0].Name())
	if err != nil {
		return nil, nil, nil, nil, 0, 0, errors.New("unable to run remote command: " + err.Error())
	}

	var fetchResp synccommon.FetchResp
	err = json.Unmarshal(output, &fetchResp)
	if err != nil {
		fmt.Println(string(output))
		return nil, nil, nil, nil, 0, 0, errors.New("unable to unmarshal server fetch response: " + err.Error())
	}
	if fetchResp.ErrMsg != nil {
		return nil, nil, nil, nil, 0, 0, errors.New("unable to complete fetch; server-side error occurred: " + strings.ReplaceAll(*fetchResp.ErrMsg, global.FSSpace, "\n"))
	}

	entryModMap := make(map[string]int64)
	ageTimestampMap := make(map[string]int64)
	var folders []string
	for folderName, containedEntries := range fetchResp.FoldersToEntries {
		folders = append(folders, folderName)
		for _, entry := range containedEntries {
			entryModMap[entry.VanityPath] = entry.ModTime
			if entry.AgeTimestamp != nil {
				ageTimestampMap[entry.VanityPath] = *entry.AgeTimestamp
			}
		}
	}

	return entryModMap, ageTimestampMap, folders, fetchResp.Deletions, fetchResp.ServerTime, clientTime, nil
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

// getRealPathSFTP formats the vanityPath to match the remote server's entry/age file directory and path separator.
func getRealPathSFTP(vanityPath, serverEntryRoot string, serverIsWindows bool) string {
	if !serverIsWindows {
		return serverEntryRoot + vanityPath
	}
	return serverEntryRoot + strings.ReplaceAll(vanityPath, "/", "\\")
}

// getRealPathSFTP formats the vanityPath to match the remote server's entry/age file directory and path separator.
func getRealAgePathSFTP(vanityPath, serverAgeDir string, serverIsWindows bool) string {
	if !serverIsWindows {
		return serverAgeDir + "/" + strings.ReplaceAll(vanityPath, "/", global.FSPath)
	}
	return serverAgeDir + "\\" + strings.ReplaceAll(vanityPath, "/", global.FSPath)
}

// sftpSync takes two slices of entries (one for downloads and one for uploads) and syncs them between the client and server using SFTP.
func sftpSync(sshClient *ssh.Client, sshEntryRoot, sshAgeDir string, sshIsWindows bool, downloadList, uploadList []string) error {
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
	for _, vanityPath := range downloadList {
		// determine if remote file is an age file
		var isAgeFile bool
		if strings.HasPrefix(vanityPath, global.FSMisc) {
			vanityPath = strings.TrimLeft(vanityPath, global.FSMisc)
			isAgeFile = true
		} else {
			filesTransferred = true // set a flag to indicate that files have been downloaded (used to determine whether to print a gap between download and upload messages)
			fmt.Println("Downloading " + back.AnsiGreen + vanityPath + back.AnsiReset)
		}

		// store path to remote entry
		var remoteFileRealPath string
		if isAgeFile {
			remoteFileRealPath = getRealAgePathSFTP(vanityPath, sshAgeDir, sshIsWindows)
		} else {
			remoteFileRealPath = getRealPathSFTP(vanityPath, sshEntryRoot, sshIsWindows)
		}

		// save modification time of remote file
		var fileInfo os.FileInfo
		fileInfo, err = sftpClient.Stat(remoteFileRealPath)
		if err != nil {
			return errors.New("unable to get remote file info (mod time): " + err.Error())
		}
		modTime := fileInfo.ModTime()

		// open remote file
		var remoteFile *sftp.File
		remoteFile, err = sftpClient.Open(remoteFileRealPath)
		if err != nil {
			return errors.New("unable to open remote file: " + err.Error())
		}

		// store path to local file
		var localFileRealPath string
		if isAgeFile {
			localFileRealPath = global.GetRealAgePath(vanityPath)
		} else {
			localFileRealPath = global.GetRealPath(vanityPath)
		}

		// create local file
		var localFile *os.File
		localFile, err = os.OpenFile(localFileRealPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
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
		err = os.Chtimes(localFileRealPath, time.Now(), modTime)
		if err != nil {
			return errors.New("unable to set local file modification time: " + err.Error())
		}
	}

	if filesTransferred {
		fmt.Println() // add a gap between download and upload messages
	}

	// iterate over the upload list
	filesTransferred = false
	for _, vanityPath := range uploadList {
		// determine if local file is an age file
		var isAgeFile bool
		if strings.HasPrefix(vanityPath, global.FSMisc) {
			vanityPath = strings.TrimLeft(vanityPath, global.FSMisc)
			isAgeFile = true
		} else {
			filesTransferred = true // set a flag to indicate that files have been uploaded (used to determine whether to print a gap between upload and sync complete messages)
			fmt.Println("Uploading " + back.AnsiBlue + vanityPath + back.AnsiReset)
		}

		// store path to local file
		var localFileRealPath string
		if isAgeFile {
			localFileRealPath = global.GetRealAgePath(vanityPath)
		} else {
			localFileRealPath = global.GetRealPath(vanityPath)
		}

		// save modification time of local file
		var fileInfo os.FileInfo
		fileInfo, err = os.Stat(localFileRealPath)
		if err != nil {
			return errors.New("unable to get local file info (mod time): " + err.Error())
		}
		modTime := fileInfo.ModTime()

		// open local file
		var localFile *os.File
		localFile, err = os.Open(localFileRealPath)
		if err != nil {
			return errors.New("unable to open local file: " + err.Error())
		}

		// store path to remote entry
		var remoteFileRealPath string
		if isAgeFile {
			remoteFileRealPath = getRealAgePathSFTP(vanityPath, sshAgeDir, sshIsWindows)
		} else {
			remoteFileRealPath = getRealPathSFTP(vanityPath, sshEntryRoot, sshIsWindows)
		}

		// create remote file
		var remoteFile *sftp.File
		remoteFile, err = sftpClient.OpenFile(remoteFileRealPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY)
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
		err = sftpClient.Chmod(remoteFileRealPath, 0600)
		if err != nil {
			return errors.New("unable to set permissions on remote file: " + err.Error())
		}

		// set the modification time of the remote file to match the value saved from the local file (from before the upload)
		err = sftpClient.Chtimes(remoteFileRealPath, time.Now(), modTime)
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
func syncLists(sshClient *ssh.Client, sshEntryRoot, sshAgeDir string, sshIsWindows, timeSynced, returnLists bool, localEntryModMap, remoteEntryModMap, localAgeTimestampMap, remoteAgeTimestampMap map[string]int64) ([3][]string, error) {
	// initialize slices to store entries that need to be downloaded or uploaded
	var downloadList, uploadList []string

	// iterate over client entries
	localMapIter := func(localMap, remoteMap map[string]int64, forAging bool) {
		for file, localTime := range localMap {
			// check if the entry is present in the server map
			if remoteTime, present := remoteMap[file]; present {
				// entry exists on both client and server, compare mod times
				if remoteTime > localTime {
					if !forAging {
						fmt.Println(back.AnsiGreen+file+back.AnsiReset, "is newer on server, adding to download list")
						downloadList = append(downloadList, file)
					} else {
						downloadList = append(downloadList, global.FSMisc+file)
					}
				} else if remoteTime < localTime {
					if !forAging {
						fmt.Println(back.AnsiBlue+file+back.AnsiReset, "is newer on client, adding to upload list")
						uploadList = append(uploadList, file)
					} else {
						uploadList = append(uploadList, global.FSMisc+file)
					}
				}
				// remove entry from remoteMap (process of elimination)
				delete(remoteMap, file)
			} else {
				if !forAging {
					fmt.Println(back.AnsiBlue+file+back.AnsiReset, "does not exist on server, adding to upload list")
					uploadList = append(uploadList, file)
				} else {
					uploadList = append(uploadList, global.FSMisc+file)
				}
			}
		}
	}
	localMapIter(localEntryModMap, remoteEntryModMap, false)
	localMapIter(localAgeTimestampMap, remoteAgeTimestampMap, true)

	// iterate over remaining entries in remote maps
	for entry := range remoteEntryModMap {
		fmt.Println(back.AnsiGreen+entry+back.AnsiReset, "does not exist on client, adding to download list")
		downloadList = append(downloadList, entry)
	}
	for ageFile := range remoteAgeTimestampMap {
		downloadList = append(downloadList, global.FSMisc+ageFile)
	}

	// call sftpSync with the download and upload lists
	if timeSynced && (max(len(downloadList), len(uploadList)) > 0) { // only call sftpSync if there are entries to download or upload
		fmt.Println() // add a gap between list-add messages and the actual sync messages from sftpSync
		err := sftpSync(sshClient, sshEntryRoot, sshAgeDir, sshIsWindows, downloadList, uploadList)
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
func deletionSync(deletions []synccommon.Deletion) error {
	var entryDeleted bool
	for _, deletion := range deletions {
		if !deletion.IsAgeFile {
			entryDeleted = true // set a flag to indicate that at least one entry has been deleted (used to determine whether to print a gap between deletion and other messages)
			fmt.Println(synccommon.AnsiDelete+deletion.VanityPath+back.AnsiReset, "has been sheared, removing locally (if it exists)")
		}
		err := os.RemoveAll(global.GetRealPath(deletion.VanityPath))
		if err != nil {
			if !deletion.IsAgeFile {
				return errors.New("unable to shear " + deletion.VanityPath + " locally: " + err.Error())
			}
			return errors.New("unable to shear age file for " + deletion.VanityPath + " locally: " + err.Error())
		}
	}
	if entryDeleted {
		fmt.Println() // add a gap between deletion and other messages
	}
	return nil
}

// folderSync creates folders on the client (from the given list of folder names).
func folderSync(folders []string) error {
	for _, folder := range folders {
		// store the full local path of the folder
		folderFullPath := global.GetRealPath(folder)

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
	sshClient, offlineMode, sshIsWindows, sshEntryRoot, sshAgeDir, err := GetSSHClient()
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
	remoteEntryModMap, remoteAgeTimestampMap, remoteFolders, deletions, serverTime, clientTime, err := getRemoteDataFromClient(sshClient)
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
	var localAgeTimestampMap map[string]int64
	localAgeTimestampMap, err = synccommon.GetEntryAges()
	if err != nil {
		return [3][]string{nil, nil, nil}, err
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
		lists, err = syncLists(sshClient, *sshEntryRoot, *sshAgeDir, *sshIsWindows, timeSynced, true, localEntryModMap, remoteEntryModMap, localAgeTimestampMap, remoteAgeTimestampMap)
		if err != nil {
			return [3][]string{nil, nil, nil}, errors.New("unable to sync entries: " + err.Error())
		}
		lists[0] = []string{} // initialize deletions list
		for _, deletion := range deletions {
			if !deletion.IsAgeFile {
				lists[0] = append(lists[0], deletion.VanityPath)
			}
		}
		return lists, nil
	}
	_, err = syncLists(sshClient, *sshEntryRoot, *sshAgeDir, *sshIsWindows, timeSynced, false, localEntryModMap, remoteEntryModMap, localAgeTimestampMap, remoteAgeTimestampMap)
	if err != nil {
		return [3][]string{nil, nil, nil}, errors.New("unable to sync entries: " + err.Error())
	}
	_ = sshClient.Close() // ignore error; non-critical/unlikely/not much could be done about it
	return lists, nil     // dummy return for when not returning lists
}
