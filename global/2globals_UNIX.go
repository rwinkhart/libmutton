//go:build !windows && !ios && !android

package global

import "github.com/rwinkhart/go-boilerplate/back"

var (
	EntryRoot = back.Home + "/.local/share/libmutton" // Path to libmutton entry directory
	CfgDir    = back.Home + "/.config/libmutton"      // Path to libmutton configuration directory
	CfgPath   = CfgDir + "/libmuttoncfg.json"         // Path to libmutton configuration file
	AgeDir    = CfgDir + "/age"                       // Path to libmutton password age directory
	SSHDir    = back.Home + "/.ssh"                   // Path to SSH directory
)

const (
	PathSeparator = "/"   // Platform-specific path separator
	IsWindows     = false // Platform indicator
)
