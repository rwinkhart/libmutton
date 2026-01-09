package core

import (
	"cmp"
	"errors"
	"maps"
	"strings"

	"github.com/rwinkhart/go-boilerplate/back"
	"github.com/rwinkhart/libmutton/config"
	"github.com/rwinkhart/libmutton/global"
	"github.com/rwinkhart/libmutton/syncclient"
	"github.com/rwinkhart/rcw/wrappers"
)

// LibmuttonInit creates the libmutton config structure based on user input.
// deviceIDPrefix can be left blank to use the system hostname.
// clientSpecificCfg can be left nil if not needed.
func LibmuttonInit(inputCB func(prompt string) string, rcwPassword []byte, appendMode, forceOfflineMode bool, deviceIDPrefix string, clientSpecificCfg map[string]any) error {
	// handle clientSpecificCfg
	newCfg := &config.CfgT{}
	if clientSpecificCfg != nil {
		newClientSpecificMap := make(map[string]any)
		maps.Copy(newClientSpecificMap, clientSpecificCfg)
		newCfg.ClientSpecific = &newClientSpecificMap
	}

	var r string
	if !forceOfflineMode {
		r = strings.ToLower(inputCB("Configure SSH settings (for synchronization)? (Y/n)"))
	} else {
		r = "n"
	}
	if len(r) > 0 && r[0] == 'n' {
		// initialize libmutton directories
		_, err := global.DirInit(appendMode)
		if err != nil {
			return errors.New("unable to initialize libmutton directories: " + err.Error())
		}

		// write config file
		offlineMode := true
		newCfg.Libmutton.OfflineMode = &offlineMode
		if err = config.Write(newCfg, false); err != nil {
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
		oldDeviceID, err := global.DirInit(appendMode)
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
		if err = config.Write(newCfg, appendMode); err != nil { // pass appendMode to allow not completely destroying existing (client-specific) config
			return err
		}
		// generate and register device ID
		sshEntryRoot, sshAgeDir, sshIsWindows, err := syncclient.GenDeviceID(oldDeviceID, deviceIDPrefix)
		if err != nil {
			return errors.New("unable to generate device ID: " + err.Error())
		}
		// update config file
		newCfg.Libmutton.SSHEntryRootPath = &sshEntryRoot
		newCfg.Libmutton.SSHAgeDirPath = &sshAgeDir
		newCfg.Libmutton.SSHIsWindows = &sshIsWindows
		if err = config.Write(newCfg, true); err != nil {
			return err
		}
	}
	// generate rcw sanity check file
	if err := RCWSanityCheckGen(rcwPassword); err != nil {
		return err
	}
	return nil
}

// RCWSanityCheckGen generates the RCW sanity check file for libmutton.
func RCWSanityCheckGen(password []byte) error {
	if err := wrappers.GenSanityCheck(global.CfgDir+global.PathSeparator+"sanity.rcw", password); err != nil {
		return errors.New("unable to generate sanity check file: " + err.Error())
	}
	return nil
}
