//go:build !windows

package global

// TargetLocationFormat returns the full location of an entry (given the name) formatted for the current platform.
func TargetLocationFormat(targetLocationIncomplete string) string {
	return EntryRoot + targetLocationIncomplete
}
