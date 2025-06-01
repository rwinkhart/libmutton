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
// rcwPassphrase and clientSpecificIniData can be left blank if not needed.
func LibmuttonInit(inputCB func(prompt string) string, clientSpecificIniData [][3]string, rcwPassphrase []byte, preserveOldConfigDir bool) error {
	r := strings.ToLower(inputCB("Configure SSH settings (for synchronization)? (y/N)"))
	if len(r) > 0 && r[0] == 'y' {
		// ensure ssh key file exists
		fallbackSSHKey := back.Home + global.PathSeparator + ".ssh" + global.PathSeparator + "id_ed25519"
		sshKeyPath := cmp.Or(back.ExpandPathWithHome(inputCB(back.AnsiBold+"Note:"+back.AnsiReset+" Only key-based authentication is supported (keys may optionally be passphrase-protected).\n      The remote server must already be in your ~"+global.PathSeparator+".ssh"+global.PathSeparator+"known_hosts file.\n\nSSH private identity file path (falls back to \""+fallbackSSHKey+"\"):")), fallbackSSHKey)
		sshKeyIsFile, _, _ := back.TargetIsFile(sshKeyPath, false, 0) // error is ignored because errorOnFail is false
		if !sshKeyIsFile {
			return errors.New("SSH identity file not found: " + sshKeyPath)
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
		oldDeviceID, err := global.DirInit(preserveOldConfigDir)
		if err != nil {
			return errors.New("unable to initialize libmutton directories: " + err.Error())
		}
		//// write config file
		//// temporarily assign sshEntryRoot and sshIsWindows to null to pass initial device ID registration
		err = cfg.WriteConfig(append(
			clientSpecificIniData,
			[][3]string{
				{"LIBMUTTON", "offlineMode", "false"},
				{"LIBMUTTON", "sshUser", sshUser},
				{"LIBMUTTON", "sshIP", sshIP},
				{"LIBMUTTON", "sshPort", sshPort},
				{"LIBMUTTON", "sshKey", sshKeyPath},
				{"LIBMUTTON", "sshKeyProtected", strconv.FormatBool(sshKeyProtected)},
				{"LIBMUTTON", "sshEntryRoot", "null"},
				{"LIBMUTTON", "sshIsWindows", "false"}}...), nil, false)
		if err != nil {
			return errors.New("unable to write config file: " + err.Error())
		}
		// generate and register device ID
		sshEntryRoot, sshIsWindows, err := synccycles.DeviceIDGen(oldDeviceID)
		if err != nil {
			return errors.New("unable to generate device ID: " + err.Error())
		}
		err = cfg.WriteConfig([][3]string{{"LIBMUTTON", "sshEntryRoot", sshEntryRoot}, {"LIBMUTTON", "sshIsWindows", sshIsWindows}}, nil, true)
		if err != nil {
			return errors.New("unable to write config file: " + err.Error())
		}
	} else {
		// initialize libmutton directories
		_, err := global.DirInit(preserveOldConfigDir)
		if err != nil {
			return errors.New("unable to initialize libmutton directories: " + err.Error())
		}
		// write config file
		if len(clientSpecificIniData) > 0 { // TODO test passing empty clientSpecificIniData
			err = cfg.WriteConfig(append(clientSpecificIniData, [][3]string{{"LIBMUTTON", "offlineMode", "true"}}...), nil, false)
			if err != nil {
				return errors.New("unable to write config file: " + err.Error())
			}
		}
	}
	// generate rcw sanity check file (if requested)
	if len(rcwPassphrase) > 0 {
		err := RCWSanityCheckGen(rcwPassphrase)
		if err != nil {
			return err
		}
	}
	return nil
}

// RCWSanityCheckGen generates the RCW sanity check file for libmutton.
func RCWSanityCheckGen(passphrase []byte) error {
	err := wrappers.GenSanityCheck(global.ConfigDir+global.PathSeparator+"sanity.rcw", passphrase)
	if err != nil {
		return errors.New("unable to generate sanity check file: " + err.Error())
	}
	return nil
}
