package synccycles

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"strings"
)

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
