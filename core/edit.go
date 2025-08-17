package core

import (
	"errors"

	"github.com/rwinkhart/go-boilerplate/back"
	"github.com/rwinkhart/libmutton/crypt"
)

// GetOldEntryData decrypts and returns old entry data (with all required lines present).
// This is a wrapper around the DecryptFileToSlice function that ensures all required lines are present in the returned slice.
// This makes it ideal for editing entries, as it guarantees at least a baseline slice length.
func GetOldEntryData(targetLocation string, field int) ([]string, error) {
	// ensure targetLocation exists and is a file
	_, err := back.TargetIsFile(targetLocation, true)
	if err != nil {
		return nil, err
	}

	// read old entry data
	decryptedEntry, err := crypt.DecryptFileToSlice(targetLocation)
	if err != nil {
		return nil, errors.New("unable to decrypt entry: " + err.Error())
	}

	// return the old entry data with all required lines present
	if field > 0 {
		return ensureSliceLength(decryptedEntry, field), nil
	} else {
		return decryptedEntry, nil
	}
}

// ensureSliceLength is a utility function that ensures a slice is long enough to contain the specified index.
func ensureSliceLength(slice []string, index int) []string {
	for len(slice) <= index {
		slice = append(slice, "")
	}
	return slice
}
