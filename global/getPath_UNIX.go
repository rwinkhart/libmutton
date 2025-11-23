//go:build !windows

package global

// GetRealPath returns the full path to an entry (given the vanity path).
func GetRealPath(vanityPath string) string {
	return EntryRoot + vanityPath
}

// GetVanityPath returns a vanityPath given a realPath
func GetVanityPath(realPath string) string {
	return realPath[rootLength:]
}
