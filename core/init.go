package core

import (
	"cmp"
	"errors"
	"maps"
	"strings"

	"github.com/rwinkhart/go-boilerplate/back"
	"github.com/rwinkhart/libmutton/cfg"
	"github.com/rwinkhart/libmutton/global"
	"github.com/rwinkhart/libmutton/syncclient"
	"github.com/rwinkhart/rcw/wrappers"
)

// LibmuttonInit creates the libmutton config structure based on user input.
// rcwPassword and clientSpecificIniData can be left blank if not needed.
func LibmuttonInit(inputCB func(prompt string) string, clientSpecificIniData map[string]any, rcwPassword []byte, preserveOldConfigDir bool, forceOfflineMode bool) error {
	// handle clientSpecificIniData
	newCfg := &cfg.ConfigT{}
	if clientSpecificIniData != nil {
		newThirdPartyMap := make(map[string]any)
		maps.Copy(newThirdPartyMap, clientSpecificIniData)
		newCfg.ThirdParty = &newThirdPartyMap
	}

	var r string
	if !forceOfflineMode {
		r = strings.ToLower(inputCB("Configure SSH settings (for synchronization)? (Y/n)"))
	} else {
		r = "n"
	}
	if len(r) > 0 && r[0] == 'n' {
		// initialize libmutton directories
		_, err := global.DirInit(preserveOldConfigDir)
		if err != nil {
			return errors.New("unable to initialize libmutton directories: " + err.Error())
		}

		// write config file
		offlineMode := true
		newCfg.Libmutton.OfflineMode = &offlineMode
		err = cfg.WriteConfig(newCfg, false)
		if err != nil {
			return err
		}
	} else {
		// ensure ssh key file exists (and is a file)
		fallbackSSHKey := back.Home + global.PathSeparator + ".ssh" + global.PathSeparator + "id_ed25519"
		sshKeyPath := cmp.Or(back.ExpandPathWithHome(inputCB(back.AnsiBold+"Note:"+back.AnsiReset+" Only key-based authentication is supported (keys may optionally be password-protected).\n      The remote server must already be in your ~"+global.PathSeparator+".ssh"+global.PathSeparator+"known_hosts file.\n\nSSH private identity file path (falls back to \""+fallbackSSHKey+"\"):")), fallbackSSHKey)
		_, err := back.TargetIsFile(sshKeyPath, true)
		if err != nil {
			return errors.New("unable to find SSH identity file: " + err.Error())
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
		//// temporarily leave sshEntryRootPath, sshAgeDirPath, and sshIsWindows as nil to pass initial device ID registration
		newCfg.Libmutton.OfflineMode = &forceOfflineMode // forceOfflineMode must be false to reach this point, so we can avoid the extra declaration
		newCfg.Libmutton.SSHUser = &sshUser
		newCfg.Libmutton.SSHIP = &sshIP
		newCfg.Libmutton.SSHPort = &sshPort
		newCfg.Libmutton.SSHKeyPath = &sshKeyPath
		newCfg.Libmutton.SSHKeyProtected = &sshKeyProtected
		err = cfg.WriteConfig(newCfg, false)
		if err != nil {
			return err
		}
		// generate and register device ID
		sshEntryRoot, sshAgeDir, sshIsWindows, err := syncclient.GenDeviceID(oldDeviceID, "")
		if err != nil {
			return errors.New("unable to generate device ID: " + err.Error())
		}
		// update config file
		newCfg.Libmutton.SSHEntryRootPath = &sshEntryRoot
		newCfg.Libmutton.SSHAgeDirPath = &sshAgeDir
		newCfg.Libmutton.SSHIsWindows = &sshIsWindows
		err = cfg.WriteConfig(newCfg, true)
		if err != nil {
			return err
		}
	}
	// generate rcw sanity check file (if requested)
	if len(rcwPassword) > 0 {
		err := RCWSanityCheckGen(rcwPassword)
		if err != nil {
			return err
		}
	}
	return nil
}

// RCWSanityCheckGen generates the RCW sanity check file for libmutton.
func RCWSanityCheckGen(password []byte) error {
	err := wrappers.GenSanityCheck(global.ConfigDir+global.PathSeparator+"sanity.rcw", password)
	if err != nil {
		return errors.New("unable to generate sanity check file: " + err.Error())
	}
	return nil
}
