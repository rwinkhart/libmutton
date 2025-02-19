package core

import (
	"os/exec"
	"strings"
)

// TODO GPG support is a temporary feature - It will be replaced with a different encryption scheme in the future

// DecryptGPG decrypts a GPG-encrypted file and returns the contents as a slice of (trimmed) strings.
func DecryptGPG(targetLocation string) []string {
	cmd := exec.Command("gpg", "--pinentry-mode", "loopback", "-q", "-d", targetLocation)
	output, err := cmd.Output()

	// ensure ANSI escape sequences are interpreted properly on Windows
	enableVirtualTerminalProcessing()

	if err != nil {
		PrintError("Failed to decrypt \""+targetLocation+"\" - Ensure it is a valid GPG-encrypted file and that you entered your passphrase correctly", ErrorDecryption, true)
	}

	return strings.Split(string(output), "\n")
}

// EncryptGPG encrypts a slice of strings using GPG and returns the encrypted data as a byte slice.
func EncryptGPG(input []string) []byte {
	gpgCfg, _ := ParseConfig([][2]string{{"LIBMUTTON", "gpgID"}}, "")
	cmd := exec.Command("gpg", "-q", "-r", gpgCfg[0], "-e")
	WriteToStdin(cmd, strings.Join(input, "\n"))
	encryptedBytes, err := cmd.Output()
	if err != nil {
		PrintError("Failed to encrypt data - Ensure that you have a valid GPG ID set in libmutton.ini", ErrorEncryption, true)
	}
	return encryptedBytes
}
