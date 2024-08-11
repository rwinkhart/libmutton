//go:build windows

package core

import "strings"

// TargetLocationFormat returns the full location of an entry (given the name) formatted for the current platform.
func TargetLocationFormat(targetLocationIncomplete string) string {
	return EntryRoot + strings.ReplaceAll(targetLocationIncomplete, "/", PathSeparator)
}
