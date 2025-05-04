package core

import (
	"os"
	"strings"

	"github.com/rwinkhart/rcw/wrappers"
)

// DecryptFileToSlice decrypts an RCW wrapped file and returns the contents as a slice of (trimmed) strings.
func DecryptFileToSlice(targetLocation string, passphrase []byte) []string {
	encBytes, err := os.ReadFile(targetLocation)
	if err != nil {
		PrintError("Failed to decrypt \""+targetLocation+"\" - "+err.Error(), ErrorDecryption, true)
	}
	decBytes, err := wrappers.Decrypt(encBytes, passphrase)
	if err != nil {
		PrintError("Failed to decrypt \""+targetLocation+"\" - "+err.Error(), ErrorDecryption, true)
	}
	return strings.Split(string(decBytes), "\n")
}

// EncryptBytes encrypts a byte slice using RCW and returns the encrypted data.
func EncryptBytes(decBytes []byte, passphrase []byte) []byte {
	encBytes := wrappers.Encrypt(decBytes, passphrase)
	return encBytes
}
