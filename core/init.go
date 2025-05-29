package core

import (
	"cmp"
	"errors"
	"strconv"
	"strings"

	"github.com/rwinkhart/go-boilerplate/back"
	"github.com/rwinkhart/libmutton/cfg"
	"github.com/rwinkhart/libmutton/global"
	"github.com/rwinkhart/libmutton/synccycles"
	"github.com/rwinkhart/rcw/wrappers"
)

// LibmuttonInit creates the libmutton config structure based on user input.
// rcwPassphrase and clientSpecificIniData are cab be left blank if not needed.
func LibmuttonInit(inputCB func(prompt string) string, clientSpecificIniData [][3]string, rcwPassphrase []byte, preserveOldConfigDir bool) error {
	r := strings.ToLower(inputCB("Configure SSH settings (for synchronization)? (y/N)"))
	if len(r) > 0 && r[0] == 'y' {
		// ensure ssh key file exists
		fallbackSSHKey := back.Home + global.PathSeparator + ".ssh" + global.PathSeparator + "id_ed25519"
		sshKeyPath := cmp.Or(back.ExpandPathWithHome(inputCB(back.AnsiBold+"Note:"+back.AnsiReset+" Only key-based authentication is supported (keys may optionally be passphrase-protected).\n      The remote server must already be in your ~"+global.PathSeparator+".ssh"+global.PathSeparator+"known_hosts file.\n\nSSH private identity file path (falls back to \""+fallbackSSHKey+"\"):")), fallbackSSHKey)
		sshKeyIsFile, _ := back.TargetIsFile(sshKeyPath, false, 0)
		if !sshKeyIsFile {
			return errors.New("ssh identity file not found: " + sshKeyPath)
		}

		// get other ssh info from user
		var sshKeyProtected bool
		r = strings.ToLower(inputCB("Is the identity file password-protected? (y/N)"))
		if len(r) > 0 && r[0] == 'y' {
			sshKeyProtected = true
		}
		sshUser := inputCB("Remote SSH username:")
		sshIP := inputCB("Remote SSH IP/domain:")
		sshPort := inputCB("Remote SSH port:")

		// perform operations based on collected user input
		//// initialize libmutton directories
		oldDeviceID := global.DirInit(preserveOldConfigDir)
		//// write config file
		//// temporarily assign sshEntryRoot and sshIsWindows to null to pass initial device ID registration
		cfg.WriteConfig(append(
			clientSpecificIniData,
			[][3]string{
				{"LIBMUTTON", "sshUser", sshUser},
				{"LIBMUTTON", "sshIP", sshIP},
				{"LIBMUTTON", "sshPort", sshPort},
				{"LIBMUTTON", "sshKey", sshKeyPath},
				{"LIBMUTTON", "sshKeyProtected", strconv.FormatBool(sshKeyProtected)},
				{"LIBMUTTON", "sshEntryRoot", "null"},
				{"LIBMUTTON", "sshIsWindows", "false"}}...), nil, false)
		// generate and register device ID
		sshEntryRoot, sshIsWindows, err := synccycles.DeviceIDGen(oldDeviceID)
		if err != nil {
			return errors.New("failed to generate device ID: " + err.Error())
		}
		cfg.WriteConfig([][3]string{{"LIBMUTTON", "sshEntryRoot", sshEntryRoot}, {"LIBMUTTON", "sshIsWindows", sshIsWindows}}, nil, true)
	} else {
		// initialize libmutton directories
		global.DirInit(preserveOldConfigDir)
		// write config file
		if len(clientSpecificIniData) > 0 { // TODO test passing empty clientSpecificIniData
			cfg.WriteConfig(clientSpecificIniData, nil, false)
		}
	}
	// generate rcw sanity check file (if requested)
	if len(rcwPassphrase) > 0 {
		err := wrappers.GenSanityCheck(global.ConfigDir+global.PathSeparator+"sanity.rcw", rcwPassphrase)
		if err != nil {
			return errors.New("failed to generate sanity check file: " + err.Error())
		}
	}
	return nil
}
