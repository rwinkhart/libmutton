package core

import (
	"errors"

	"github.com/rwinkhart/go-boilerplate/back"
	"github.com/rwinkhart/libmutton/crypt"
)

// GetOldEntryData decrypts and returns old entry data (with all required lines present).
func GetOldEntryData(targetLocation string, field int) ([]string, error) {
	// ensure targetLocation exists
	back.TargetIsFile(targetLocation, true, 2)

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
