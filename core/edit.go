package core

import "github.com/rwinkhart/go-boilerplate/back"

// GetOldEntryData decrypts and returns old entry data (with all required lines present).
func GetOldEntryData(targetLocation string, field int) []string {
	// ensure targetLocation exists
	back.TargetIsFile(targetLocation, true, 2)

	// read old entry data
	unencryptedEntry := DecryptFileToSlice(targetLocation)

	// return the old entry data with all required lines present
	if field > 0 {
		return ensureSliceLength(unencryptedEntry, field)
	} else {
		return unencryptedEntry
	}
}

// ensureSliceLength is a utility function that ensures a slice is long enough to contain the specified index.
func ensureSliceLength(slice []string, index int) []string {
	for len(slice) <= index {
		slice = append(slice, "")
	}
	return slice
}
