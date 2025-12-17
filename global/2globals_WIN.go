//go:build windows

package global

import "github.com/rwinkhart/go-boilerplate/back"

var (
	EntryRoot  = back.Home + "\\AppData\\Local\\libmutton\\entries" // Path to libmutton entry directory
	ConfigDir  = back.Home + "\\AppData\\Local\\libmutton\\config"  // Path to libmutton configuration directory
	ConfigPath = ConfigDir + "\\libmuttoncfg.json"                  // Path to libmutton configuration file
	AgeDir     = ConfigDir + "\\age"                                // Path to libmutton password age directory
)

const (
	PathSeparator = "\\" // Platform-specific path separator
	IsWindows     = true // Platform indicator
)
