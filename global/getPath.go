package global

import "strings"

// GetRealAgePath returns the full path to an age file (given the vanity path)
func GetRealAgePath(vanityPath string) string {
	return AgeDir + PathSeparator + strings.ReplaceAll(vanityPath, "/", FSPath)
}
