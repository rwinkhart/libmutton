//go:build ios

package global

import (
	"path/filepath"

	"github.com/rwinkhart/go-boilerplate/back"
)

func init() {
	back.Home = fruityHome
}

var (
	fruityHome = filepath.Join(back.Home, "Library", "Application Support", "libmutton-ios") // private, non-iCloud-backed dir
	EntryRoot  = fruityHome + "/entries"                                                     // Path to libmutton entry directory
	CfgDir     = fruityHome + "/config"                                                      // Path to libmutton configuration directory
	CfgPath    = CfgDir + "/libmuttoncfg.json"                                               // Path to libmutton configuration file
	AgeDir     = CfgDir + "/age"                                                             // Path to libmutton password age directory
	SSHDir     = fruityHome + "/ssh"                                                         // Path to SSH directory
)

const (
	PathSeparator = "/"   // Platform-specific path separator
	IsWindows     = false // Platform indicator
)
