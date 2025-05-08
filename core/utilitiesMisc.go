package core

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"io"
	"math"
	"math/big"
	"os"
	"os/exec"
	"strings"
)

// TargetIsFile checks if the targetLocation is a file, directory, or is inaccessible.
// Requires: failCondition (0 = fail on inaccessible, 1 = fail on inaccessible&file, 2 = fail on inaccessible&directory).
// Returns: isFile, isAccessible.
func TargetIsFile(targetLocation string, errorOnFail bool, failCondition uint8) (bool, bool) {
	targetInfo, err := os.Stat(targetLocation)
	if err != nil {
		if errorOnFail {
			PrintError("Failed to access \""+targetLocation+"\" - Ensure it exists and has the correct permissions", ErrorTargetNotFound, true)
		}
		return false, false
	}
	if targetInfo.IsDir() {
		if errorOnFail && failCondition == 2 {
			PrintError("\""+targetLocation+"\" is a directory", ErrorTargetWrongType, true)
		}
		return false, true
	} else {
		if errorOnFail && failCondition == 1 {
			PrintError("\""+targetLocation+"\" is a file", ErrorTargetWrongType, true)
		}
		return true, true
	}
}

// WriteEntry writes entryData to an encrypted file at targetLocation.
func WriteEntry(targetLocation string, entryData []byte) {
	encryptedBytes := EncryptBytes(entryData)
	err := os.WriteFile(targetLocation, encryptedBytes, 0600)
	if err != nil {
		PrintError("Failed to write to file: "+err.Error(), ErrorWrite, true)
	}
}

// writeToStdin is a utility function that writes a string to a command's stdin.
func writeToStdin(cmd *exec.Cmd, input string) {
	stdin, err := cmd.StdinPipe()
	if err != nil {
		PrintError("Failed to access stdin for system command: "+err.Error(), ErrorOther, true)
	}

	go func() {
		defer func(stdin io.WriteCloser) {
			_ = stdin.Close() // error ignored; if stdin could be accessed, it can probably be closed
		}(stdin)
		_, _ = io.WriteString(stdin, input)
	}()
}

// readFromStdin is a utility function that reads a string from stdin.
func readFromStdin() string {
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text()
	}
	return ""
}

// CreateTempFile creates a temporary file and returns a pointer to it.
func CreateTempFile() *os.File {
	tempFile, err := os.CreateTemp("", "*.markdown")
	if err != nil {
		PrintError("Failed to create temporary file: "+err.Error(), ErrorWrite, true)
	}
	return tempFile
}

// RemoveTrailingEmptyStrings removes empty strings from the end of a slice.
func RemoveTrailingEmptyStrings(slice []string) []string {
	for i := len(slice) - 1; i >= 0; i-- {
		if slice[i] != "" {
			return slice[:i+1]
		}
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
func EntryAddPrecheck(targetLocation string) uint8 {
	// ensure target location does not already exist
	_, isAccessible := TargetIsFile(targetLocation, false, 0)
	if isAccessible {
		PrintError("Target location already exists", ErrorTargetExists, false)
		return 1 // inform interactive clients that the target location already exists
	}
	// ensure target containing directory exists and is a directory (not a file)
	containingDir := targetLocation[:strings.LastIndex(targetLocation, PathSeparator)]
	isFile, isAccisAccessible := TargetIsFile(containingDir, false, 1)
	if isFile || !isAccisAccessible {
		PrintError("\""+containingDir+"\" is not a valid containing directory", ErrorTargetWrongType, false)
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

// ExpandPathWithHome, given a path (as a string) containing "~", returns the path with "~" expanded to the user's home directory.
func ExpandPathWithHome(path string) string {
	return strings.Replace(path, "~", Home, 1)
}

// PrintError prints an error message in the standard libmutton format and exits with the specified exit code.
// Requires: message (the error message to print),
// exitCode (the exit code to use),
// forceHardExit (if true, exit immediately; if false, allow soft exit for interactive clients).
func PrintError(message string, exitCode int, forceHardExit bool) {
	fmt.Println(AnsiError + message + AnsiReset)
	if forceHardExit {
		os.Exit(exitCode)
	} else {
		Exit(exitCode)
	}
}
