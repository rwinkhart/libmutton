//go:build !windows

package global

import "github.com/rwinkhart/go-boilerplate/back"

var (
	EntryRoot  = back.Home + "/.local/share/libmutton" // Path to libmutton entry directory
	ConfigDir  = back.Home + "/.config/libmutton"      // Path to libmutton configuration directory
	ConfigPath = ConfigDir + "/libmuttoncfg.json"      // Path to libmutton configuration file
	AgeDir     = ConfigDir + "/age"                    // Path to libmutton password age directory
)

const (
	PathSeparator = "/"   // Platform-specific path separator
	IsWindows     = false // Platform indicator
)
