//go:build !windows

package core

var EntryRoot = Home + "/.local/share/libmutton" // path to libmutton entry directory
var ConfigDir = Home + "/.config/libmutton"      // path to libmutton configuration directory
var ConfigPath = ConfigDir + "/libmutton.ini"    // path to libmutton configuration file

const (
	PathSeparator = "/"   // platform-specific path separator
	IsWindows     = false // platform indicator
)

// enableVirtualTerminalProcessing is a dummy function on UNIX-like systems (only needed on Windows).
// TODO Remove after migration off of GPG, as pinentry is responsible for disabling ANSI escape sequence interpretation.
func enableVirtualTerminalProcessing() {
	return
}
