//go:build ios

package global

import (
	"path/filepath"

	"github.com/rwinkhart/go-boilerplate/back"
)

func init() {
	// get the app's Library/Application Support directory (private, not backed up by iCloud)
	back.Home = filepath.Join(back.Home, "Library", "Application Support", "libmutton-ios")
}

var (
	EntryRoot = back.Home + "/entries"        // Path to libmutton entry directory
	CfgDir    = back.Home + "/config"         // Path to libmutton configuration directory
	CfgPath   = CfgDir + "/libmuttoncfg.json" // Path to libmutton configuration file
	AgeDir    = CfgDir + "/age"               // Path to libmutton password age directory
	SSHDir    = back.Home + "/ssh"            // Path to SSH directory
)

const (
	PathSeparator = "/"   // Platform-specific path separator
	IsWindows     = false // Platform indicator
)
