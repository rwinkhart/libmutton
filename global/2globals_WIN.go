//go:build windows

package global

import "github.com/rwinkhart/go-boilerplate/back"

var (
	EntryRoot  = back.Home + "\\AppData\\Local\\libmutton\\entries" // Path to libmutton entry directory
	ConfigDir  = back.Home + "\\AppData\\Local\\libmutton\\config"  // Path to libmutton configuration directory
	ConfigPath = ConfigDir + "\\libmutton.ini"                      // Path to libmutton configuration file
)

const (
	PathSeparator = "\\" // Platform-specific path separator
	IsWindows     = true // Platform indicator
)
