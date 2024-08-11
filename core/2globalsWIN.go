//go:build windows

package core

import (
	"os"
	"syscall"
)

var EntryRoot = Home + "\\AppData\\Local\\libmutton\\entries" // path to libmutton entry directory
var ConfigDir = Home + "\\AppData\\Local\\libmutton\\config"  // path to libmutton configuration directory
var ConfigPath = ConfigDir + "\\libmutton.ini"                // path to libmutton configuration file

const (
	PathSeparator = "\\" // platform-specific path separator
	IsWindows     = true // platform indicator
)

// enableVirtualTerminalProcessing ensures ANSI escape sequences are interpreted properly on Windows.
// TODO Remove after migration off of GPG, as pinentry is responsible for disabling ANSI escape sequence interpretation.
func enableVirtualTerminalProcessing() {
	stdout := syscall.Handle(os.Stdout.Fd())

	var originalMode uint32
	syscall.GetConsoleMode(stdout, &originalMode)
	originalMode |= 0x0004

	syscall.MustLoadDLL("kernel32").MustFindProc("SetConsoleMode").Call(uintptr(stdout), uintptr(originalMode))
}
