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
	"github.com/rwinkhart/libmutton/age"
	"github.com/rwinkhart/libmutton/config"
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
// Only supports key-based authentication (passwords are supported for CLI-based implementations).
func GetSSHClient() (*ssh.Client, bool, *bool, *string, *string, error) {
	// get SSH config info
	cfg, err := config.Load()
	if err != nil {
		return nil, false, nil, nil, nil, errors.New("unable to parse SSH config: " + err.Error())
	}
	if *cfg.Libmutton.OfflineMode {
		return nil, true, nil, nil, nil, nil
	}

	// read private key
	key, err := os.ReadFile(*cfg.Libmutton.SSHKeyPath)
	if err != nil {
		return nil, false, nil, nil, nil, errors.New("unable to read private key: " + *cfg.Libmutton.SSHKeyPath)
	}

	// parse private key
	var parsedKey ssh.Signer
	if !*cfg.Libmutton.SSHKeyProtected {
		parsedKey, err = ssh.ParsePrivateKey(key)
	} else {
		parsedKey, err = ssh.ParsePrivateKeyWithPassphrase(key, global.GetPassword("Enter password for your SSH keyfile:"))
	}
	if err != nil {
		return nil, false, nil, nil, nil, errors.New("unable to parse private key: " + *cfg.Libmutton.SSHKeyPath)
	}

	// read known hosts file
	var hostKeyCallback ssh.HostKeyCallback
	hostKeyCallback, err = knownhosts.New(back.Home + global.PathSeparator + ".ssh" + global.PathSeparator + "known_hosts")
	if err != nil {
		return nil, false, nil, nil, nil, errors.New("unable to read known hosts file: " + err.Error())
	}

	// configure SSH client
	sshCfg := &ssh.ClientConfig{
		User: *cfg.Libmutton.SSHUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(parsedKey),
		},
		HostKeyCallback: hostKeyCallback,
		Timeout:         3 * time.Second,
	}

	// connect to SSH server
	sshClient, err := ssh.Dial("tcp", *cfg.Libmutton.SSHIP+":"+*cfg.Libmutton.SSHPort, sshCfg)
	if err != nil {
		return nil, false, nil, nil, nil, errors.New("unable to connect to remote server: " + err.Error())
	}

	return sshClient, false, cfg.Libmutton.SSHIsWindows, cfg.Libmutton.SSHEntryRootPath, cfg.Libmutton.SSHAgeDirPath, nil
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
// a map of remote vanityPaths to their containing folders and mod+age timestamps,
// a list of queued deletions,
// and the current server&client times as UNIX timestamps.
func getRemoteDataFromClient(sshClient *ssh.Client) (synccommon.EntriesMap, []synccommon.Deletion, int64, int64, error) {
	// get remote output over SSH
	deviceIDList, err := global.GenDeviceIDList()
	if err != nil {
		return nil, nil, 0, 0, err
	}
	if len(deviceIDList) == 0 {
		return nil, nil, 0, 0, errors.New("no device ID found")
	}
	clientTime := time.Now().Unix() // get client time now to avoid accuracy issues caused by unpredictable sync time
	output, err := GetSSHOutput(sshClient, "libmuttonserver fetch", (deviceIDList)[0].Name())
	if err != nil {
		return nil, nil, 0, 0, errors.New("unable to run remote command: " + err.Error())
	}

	var fetchResp synccommon.FetchResp
	if err = json.Unmarshal(output, &fetchResp); err != nil {
		fmt.Println(string(output))
		return nil, nil, 0, 0, errors.New("unable to unmarshal server fetch response: " + err.Error())
	}
	if fetchResp.ErrMsg != nil {
		return nil, nil, 0, 0, errors.New("unable to complete fetch; server-side error occurred: " + strings.ReplaceAll(*fetchResp.ErrMsg, global.FSSpace, "\n"))
	}

	return fetchResp.Entries, fetchResp.Deletions, fetchResp.ServerTime, clientTime, nil
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
		filesTransferred = true // set a flag to indicate that files have been downloaded (used to determine whether to print a gap between download and upload messages)
		fmt.Println("Downloading " + back.AnsiGreen + vanityPath + back.AnsiReset)

		// store path to remote entry
		remoteFileRealPath := getRealPathSFTP(vanityPath, sshEntryRoot, sshIsWindows)

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
		localFileRealPath := global.GetRealPath(vanityPath)

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
		if err = os.Chtimes(localFileRealPath, time.Now(), modTime); err != nil {
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
		if err = sftpClient.Chmod(remoteFileRealPath, 0600); err != nil {
			return errors.New("unable to set permissions on remote file: " + err.Error())
		}

		// set the modification time of the remote file to match the value saved from the local file (from before the upload)
		if err = sftpClient.Chtimes(remoteFileRealPath, time.Now(), modTime); err != nil {
			return errors.New("unable to set remote file modification time: " + err.Error())
		}
	}

	if filesTransferred {
		fmt.Println() // add a gap between upload and sync complete messages
	}

	return nil
}

// syncLists determines which entries need to be downloaded and uploaded
// for synchronization and calls sftpSync with this information.
func syncLists(sshClient *ssh.Client, sshEntryRoot, sshAgeDir string, sshIsWindows bool, timeSyncedErr error, localEntryMap, remoteEntryMap synccommon.EntriesMap) ([3][]string, error) {
	// initialize slices to store entries that need to be downloaded or uploaded
	var downloadList, uploadList []string

	// iterate over client entries in local map
	for vanityPath, localInfo := range localEntryMap {
		// check if the entry is present on the server
		if _, exists := remoteEntryMap[vanityPath]; exists {
			// entry exists on both client and server, compare mod times
			remoteInfo := remoteEntryMap[vanityPath]
			if remoteInfo.ModTime > localInfo.ModTime {
				fmt.Println(back.AnsiGreen+vanityPath+back.AnsiReset, "is newer on server, adding to download list")
				downloadList = append(downloadList, vanityPath)
				if remoteInfo.AgeTimestamp != nil {
					if err := age.Entry(vanityPath, *remoteInfo.AgeTimestamp); err != nil {
						return [3][]string{nil, nil, nil}, errors.New("unable to update age timestamp for " + vanityPath + ": " + err.Error())
					}
				}
			} else if remoteInfo.ModTime < localInfo.ModTime {
				fmt.Println(back.AnsiBlue+vanityPath+back.AnsiReset, "is newer on client, adding to upload list")
				uploadList = append(uploadList, vanityPath)
				if localInfo.AgeTimestamp != nil && localInfo.AgeTimestamp != remoteInfo.AgeTimestamp {
					uploadList = append(uploadList, global.FSMisc+vanityPath)
				}
			}
			// remove entry from remote map (process of elimination)
			delete(remoteEntryMap, vanityPath)
		} else {
			fmt.Println(back.AnsiBlue+vanityPath+back.AnsiReset, "does not exist on server, adding to upload list")
			uploadList = append(uploadList, vanityPath)
			if localInfo.AgeTimestamp != nil {
				uploadList = append(uploadList, global.FSMisc+vanityPath)
			}
		}
	}

	// iterate over remaining entries in remote map
	for vanityPath, remoteInfo := range remoteEntryMap {
		fmt.Println(back.AnsiGreen+vanityPath+back.AnsiReset, "does not exist on client, adding to download list")
		downloadList = append(downloadList, vanityPath)
		if err := os.MkdirAll(global.GetRealPath(remoteInfo.ContainingFolder), 0700); err != nil {
			return [3][]string{nil, nil, nil}, errors.New("unable to create containing folder for " + vanityPath + ": " + err.Error())
		}
		if remoteInfo.AgeTimestamp != nil {
			if err := age.Entry(vanityPath, *remoteInfo.AgeTimestamp); err != nil {
				return [3][]string{nil, nil, nil}, errors.New("unable to update age timestamp for " + vanityPath + ": " + err.Error())
			}
		}
	}

	// call sftpSync with the download and upload lists
	if timeSyncedErr == nil && (max(len(downloadList), len(uploadList)) > 0) { // only call sftpSync if there are entries to download or upload
		fmt.Println() // add a gap between list-add messages and the actual sync messages from sftpSync
		if err := sftpSync(sshClient, sshEntryRoot, sshAgeDir, sshIsWindows, downloadList, uploadList); err != nil {
			return [3][]string{nil, nil, nil}, errors.New("unable to sync entries: " + err.Error())
		}
		fmt.Println("Client is synchronized with server")
	}

	return [3][]string{nil, downloadList, uploadList}, timeSyncedErr
}

// deletionSync removes entries from the client that have been deleted on the server (multi-client deletion).
func deletionSync(deletions []synccommon.Deletion) error {
	var entryDeleted bool
	for _, deletion := range deletions {
		if !deletion.IsAgeFile {
			entryDeleted = true // set a flag to indicate that at least one entry has been deleted (used to determine whether to print a gap between deletion and other messages)
			fmt.Println(synccommon.AnsiDelete+deletion.VanityPath+back.AnsiReset, "has been sheared, removing locally (if it exists)")
		}
		if err := os.RemoveAll(global.GetRealPath(deletion.VanityPath)); err != nil {
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

// RunJob runs the SSH sync job and returns deletions, downloads,
// and uploads lists for the client to report to the user.
func RunJob() ([3][]string, error) {
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
	remoteEntryMap, deletions, serverTime, clientTime, err := getRemoteDataFromClient(sshClient)
	if err != nil {
		return [3][]string{nil, nil, nil}, errors.New("unable to fetch remote data: " + err.Error())
	}

	// sync deletions
	if err = deletionSync(deletions); err != nil {
		return [3][]string{nil, nil, nil}, errors.New("unable to sync deletions: " + err.Error())
	}

	// fetch local lists
	localEntryMap, err := synccommon.GetAllEntryData()
	if err != nil {
		return [3][]string{nil, nil, nil}, errors.New("unable to fetch local entry data: " + err.Error())
	}

	// before syncing lists, ensure the client and server clocks are synced within 45 seconds
	var timeSyncedErr error
	timeDiff := serverTime - clientTime
	if timeDiff < -45 || timeDiff > 45 {
		timeSyncedErr = errors.New("client and server clocks are out of sync\n\nplease ensure both clocks are correct before attempting to sync again\n\na dry sync has been performed; it is strongly recommended to review it and manually update the modification times as applicable to ensure the correct version of each entry is kept\n\nif the client's clock is at fault, update the modification times of any entries pending upload, even if the correct (upload) operation is being performed on them; failure to do so could result in entries being uploaded to the server with the incorrect modification times (could result in data loss)" + back.AnsiReset)
	}

	// sync new and updated entries
	// if time is not synced, the time sync error and upload/download lists will be returned here
	lists, err := syncLists(sshClient, *sshEntryRoot, *sshAgeDir, *sshIsWindows, timeSyncedErr, localEntryMap, remoteEntryMap)
	if err != nil {
		return lists, errors.New("unable to sync entries: " + err.Error())
	}

	// add deletions info to sync lists
	lists[0] = []string{}
	for _, deletion := range deletions {
		if !deletion.IsAgeFile {
			lists[0] = append(lists[0], deletion.VanityPath)
		}
	}

	return lists, nil
}
