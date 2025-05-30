package core

import (
	"errors"
	"os"
	"strings"

	"github.com/rwinkhart/go-boilerplate/back"
	"github.com/rwinkhart/libmutton/crypt"
	"github.com/rwinkhart/libmutton/global"
)

// WriteEntry writes entryData to an encrypted file at targetLocation.
func WriteEntry(targetLocation string, entryData []byte) error {
	encBytes := crypt.EncryptBytes(entryData)
	err := os.WriteFile(targetLocation, encBytes, 0600)
	if err != nil {
		return errors.New("unable to write to file: " + err.Error())
	}
	return nil
}

// ClampTrailingWhitespace strips trailing newlines, carriage returns, and tabs from each line in a note.
// Additionally, it removes single trailing spaces and truncates multiple trailing spaces to two (for Markdown formatting).
func ClampTrailingWhitespace(note []string) {
	for i, line := range note {
		// remove trailing tabs, carriage returns, and newlines
		note[i] = strings.TrimRight(line, "\t\r\n")

		// determine the number of trailing spaces
		var endSpacesCount int
		for j := len(line) - 1; j >= 0; j-- {
			if line[j] != ' ' {
				break
			}
			endSpacesCount++
		}

		// remove single spaces, truncate multiple spaces (leave two for Markdown formatting)
		switch endSpacesCount {
		case 0:
			// do nothing
		case 1:
			// remove the trailing space
			note[i] = strings.TrimRight(line, " ")
		default:
			// truncate the trailing spaces to two
			note[i] = line[:len(line)-endSpacesCount+2]
		}
	}
}

// EntryAddPrecheck ensures the directory meant to contain a new
// entry exists and that the target entry location is not already used.
// Returns: statusCode (0 = success, 1 = target location already exists, 2 = containing directory is invalid).
func EntryAddPrecheck(targetLocation string) (uint8, error) {
	// ensure target location does not already exist
	_, isAccessible := back.TargetIsFile(targetLocation, false, 0)
	if isAccessible {
		return 1, errors.New("target location already exists")
	}
	// ensure target containing directory exists and is a directory (not a file)
	containingDir := targetLocation[:strings.LastIndex(targetLocation, global.PathSeparator)]
	isFile, isAccessible := back.TargetIsFile(containingDir, false, 1)
	if isFile || !isAccessible {
		return 2, errors.New("\"" + containingDir + "\" is not a valid containing directory")
	}
	return 0, nil
}

// EntryIsNotEmpty iterates through entryData and returns true if any line is not empty.
func EntryIsNotEmpty(entryData []string) bool {
	for _, line := range entryData {
		if line != "" {
			return true
		}
	}
	return false
}
