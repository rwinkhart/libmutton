//go:build !windows

package core

var (
	EntryRoot  = Home + "/.local/share/libmutton" // Path to libmutton entry directory
	ConfigDir  = Home + "/.config/libmutton"      // Path to libmutton configuration directory
	ConfigPath = ConfigDir + "/libmutton.ini"     // Path to libmutton configuration file
)

const (
	PathSeparator = "/"   // Platform-specific path separator
	IsWindows     = false // Platform indicator
)

// enableVirtualTerminalProcessing is a dummy function on UNIX-like systems (only needed on Windows).
// TODO Remove after migration off of GPG, as pinentry is responsible for disabling ANSI escape sequence interpretation.
func enableVirtualTerminalProcessing() {
	return
}
