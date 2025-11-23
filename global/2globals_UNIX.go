//go:build !windows

package global

import "github.com/rwinkhart/go-boilerplate/back"

var (
	EntryRoot  = back.Home + "/.local/share/libmutton" // Path to libmutton entry directory
	ConfigDir  = back.Home + "/.config/libmutton"      // Path to libmutton configuration directory
	ConfigPath = ConfigDir + "/libmutton.ini"          // Path to libmutton configuration file
	AgeDir     = ConfigDir + "/aging"                  // Path to libmutton password aging database
)

const (
	PathSeparator = "/"   // Platform-specific path separator
	IsWindows     = false // Platform indicator
)
