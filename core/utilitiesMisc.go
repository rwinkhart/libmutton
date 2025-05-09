package core

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"os"
	"strings"

	"github.com/rwinkhart/go-boilerplate/back"
)

// WriteEntry writes entryData to an encrypted file at targetLocation.
func WriteEntry(targetLocation string, entryData []byte) {
	encryptedBytes := EncryptBytes(entryData)
	err := os.WriteFile(targetLocation, encryptedBytes, 0600)
	if err != nil {
		back.PrintError("Failed to write to file: "+err.Error(), back.ErrorWrite, true)
	}
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
func EntryAddPrecheck(targetLocation string) uint8 {
	// ensure target location does not already exist
	_, isAccessible := back.TargetIsFile(targetLocation, false, 0)
	if isAccessible {
		back.PrintError("Target location already exists", ErrorTargetExists, false)
		return 1 // inform interactive clients that the target location already exists
	}
	// ensure target containing directory exists and is a directory (not a file)
	containingDir := targetLocation[:strings.LastIndex(targetLocation, PathSeparator)]
	isFile, isAccisAccessible := back.TargetIsFile(containingDir, false, 1)
	if isFile || !isAccisAccessible {
		back.PrintError("\""+containingDir+"\" is not a valid containing directory", back.ErrorTargetWrongType, false)
		return 2 // inform interactive clients that the containing directory is invalid
	}
	return 0
}

// StringGen generates a random string of a specified length and complexity.
// Requires: complexity (minimum percentage of special characters to be returned in the generated string; set to 0 to generate a simple string),
// complexCharsetLevel (1 = safe for filenames, 2 = safe for most password entries, 3 = safe only for well-made password entries)
func StringGen(length int, complexity float64, complexCharsetLevel uint8) string {
	var actualSpecialChars int // track the number of special characters in the generated string
	var minSpecialChars int    // track the minimum number of special characters to accept
	var extendedCharset string // additions to character set used for complex strings

	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" // default character set used for all strings
	const extendedCharsetFiles = "!#$%&+,-.;=@_~^()[]{}`'"                      // additional special characters for complex strings (safe in file names)
	const extendedCharsetMostPassword = "*:><?|"                                // additional special characters for complex strings (NOT safe in file names)
	const extendedCharsetSpecialPassword = "\"/\\"                              // additional special characters for complex strings (NOT safe in file names)

	if complexity > 0 {
		minSpecialChars = int(math.Round(float64(length) * complexity)) // determine minimum number of special characters to accept
		switch complexCharsetLevel {
		case 1:
			extendedCharset = extendedCharsetFiles
		case 2:
			extendedCharset = extendedCharsetMostPassword + extendedCharsetFiles[:len(extendedCharsetFiles)-9]
		case 3:
			extendedCharset = extendedCharsetFiles + extendedCharsetMostPassword + extendedCharsetSpecialPassword
		}
		charset += extendedCharset
	}

	// loop until a string of the desired complexity is generated
	for {
		// generate a random string
		result := make([]byte, length)
		for i := range result {
			val, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
			result[i] = charset[val.Int64()]
		}

		// return early if the string is not complex
		if complexity <= 0 {
			return string(result)
		}

		// count the number of special characters in the generated string
		for _, char := range string(result) {
			if strings.ContainsRune(extendedCharset, char) {
				actualSpecialChars++
			}
		}

		// return the generated string if it contains enough special characters
		if actualSpecialChars >= minSpecialChars {
			return string(result)
		}

		// reset special character counter
		fmt.Println("Regenerating string until desired complexity is achieved...")
		actualSpecialChars = 0
	}
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
