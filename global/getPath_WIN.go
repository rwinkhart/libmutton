//go:build windows

package global

import (
	"strings"
)

// GetRealPath returns the full path to an entry (given the vanity path).
func GetRealPath(vanityPath string) string {
	return EntryRoot + strings.ReplaceAll(vanityPath, "/", PathSeparator)
}

// GetVanityPath returns a vanityPath given a realPath
func GetVanityPath(realPath string) string {
	return strings.ReplaceAll(realPath[RootLength:], "\\", "/")
}
