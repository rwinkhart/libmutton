package core

import (
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
			fmt.Println(AnsiError + "Failed to access \"" + targetLocation + "\" - Ensure it exists and has the correct permissions" + AnsiReset)
			os.Exit(ErrorTargetNotFound)
		}
		return false, false
	}
	if targetInfo.IsDir() {
		if errorOnFail && failCondition == 2 {
			fmt.Println(AnsiError + "\"" + targetLocation + "\" is a directory" + AnsiReset)
			os.Exit(ErrorTargetWrongType)
		}
		return false, true
	} else {
		if errorOnFail && failCondition == 1 {
			fmt.Println(AnsiError + "\"" + targetLocation + "\" is a file" + AnsiReset)
			os.Exit(ErrorTargetWrongType)
		}
		return true, true
	}
}

// WriteEntry writes entryData to an encrypted file at targetLocation.
func WriteEntry(targetLocation string, entryData []string) {
	encryptedBytes := EncryptGPG(entryData)
	err := os.WriteFile(targetLocation, encryptedBytes, 0600)
	if err != nil {
		fmt.Println(AnsiError+"Failed to write to file:", err.Error()+AnsiReset)
		os.Exit(ErrorWrite)
	}
}

// WriteToStdin is a utility function that writes a string to a command's stdin.
// TODO unexport (import?) after migration off of GPG
func WriteToStdin(cmd *exec.Cmd, input string) {
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Println(AnsiError+"Failed to access stdin for system command:", err.Error()+AnsiReset)
		os.Exit(ErrorOther)
	}

	go func() {
		defer func(stdin io.WriteCloser) {
			_ = stdin.Close() // error ignored; if stdin could be accessed, it can probably be closed
		}(stdin)
		_, _ = io.WriteString(stdin, input)
	}()
}

// CreateTempFile creates a temporary file and returns a pointer to it.
func CreateTempFile() *os.File {
	tempFile, err := os.CreateTemp("", "*.markdown")
	if err != nil {
		fmt.Println(AnsiError+"Failed to create temporary file:", err.Error()+AnsiReset)
		os.Exit(ErrorWrite)
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
	return []string{}
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

// StringGen generates a random string of a specified length and complexity.
// Requires: complexity (minimum percentage of special characters to be returned in the generated string; only impacts complex strings),
// safeForFileName: (if true, the generated string will only contain special characters that are safe for file names; only impacts complex strings).
func StringGen(length int, complex bool, complexity float64, safeForFileName bool) string {
	var actualSpecialChars int // track the number of special characters in the generated string
	var minSpecialChars int    // track the minimum number of special characters to accept
	var extendedCharset string // additions to character set used for complex strings

	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" // default character set used for all strings
	const extendedCharsetFiles = "!#$%&'()+,-.;=@[]^_`{}~"                      // additional special characters for complex strings (safe in file names)
	const extendedCharsetPassword = "\"*:><?/\\|"                               // additional special characters for complex strings (NOT safe in file names)
	if complex {
		minSpecialChars = int(math.Round(float64(length) * complexity)) // determine minimum number of special characters to accept
		if !safeForFileName {
			extendedCharset = extendedCharsetFiles + extendedCharsetPassword
		} else {
			extendedCharset = extendedCharsetFiles
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
		if !complex {
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
