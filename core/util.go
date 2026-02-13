package core

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/rwinkhart/go-boilerplate/back"
	"github.com/rwinkhart/libmutton/age"
	"github.com/rwinkhart/libmutton/crypt"
	"github.com/rwinkhart/libmutton/global"
	"github.com/rwinkhart/libmutton/syncclient"
	"github.com/rwinkhart/libmutton/synccommon"
	"github.com/rwinkhart/rcw/wrappers"
)

// WriteEntry writes decSlice to an encrypted file at realPath.
// If the entry contains an updated password, an age file is also created.
// Leave rcwPassword nil to use RCW demonization.
func WriteEntry(realPath string, decSlice []string, passwordIsNew bool, rcwPassword []byte) error {
	err := os.WriteFile(realPath, crypt.EncryptBytes([]byte(strings.Join(clampTrailingWhitespace(decSlice), "\n")), rcwPassword), 0600)
	if err != nil {
		return errors.New("unable to write to file: " + err.Error())
	}

	if decSlice != nil {
		if passwordIsNew { // update age data when password changes
			if decSlice[0] != "" { // if the password change was NOT a removal, update the age file
				if err = age.Entry(global.GetVanityPath(realPath), time.Now().Unix()); err != nil {
					return errors.New("unable to update age data: " + err.Error())
				}
			} else { // if the password change was a removal, remove the associated age file
				if err = syncclient.ShearRemote(global.GetVanityPath(realPath), true); err != nil {
					return errors.New("unable to remove age data: " + err.Error())
				}
			}
		}
	}

	return nil
}

// EntryRefresh re-encrypts all libmutton entries with a new password
// and optimizes each entry to ensure they are as slim as possible.
func EntryRefresh(oldRCWPassword, newRCWPassword []byte, removeOldDir bool) error {
	// ensure global.EntryRoot+"-new" and global.EntryRoot-"old" do not exist
	dirEnds := []string{"-new", "-old"}
	for i, dirEnd := range dirEnds {
		if i == 1 && !removeOldDir {
			if _, err := os.Stat(global.EntryRoot + "-old"); !os.IsNotExist(err) {
				return errors.New("unable to refresh entries: \"" + global.EntryRoot + "-old\" already exists")
			}
		}
		if err := os.RemoveAll(global.EntryRoot + dirEnd); err != nil {
			return errors.New("unable to remove \"" + global.EntryRoot + dirEnd + "\": " + err.Error())
		}
	}

	// create output directory structure (global.EntryRoot + "-new"/*)
	entries, folders, err := synccommon.WalkEntryDir()
	if err != nil {
		return errors.New("unable to walk entry directory: " + err.Error())
	}
	for _, folder := range folders {
		fullPath := global.EntryRoot + "-new" + strings.ReplaceAll(folder, "/", global.PathSeparator)
		if err := os.MkdirAll(fullPath, 0700); err != nil {
			return errors.New("unable to create temporary directory \"" + fullPath + "\": " + err.Error())
		}
	}

	// decrypt, optimize, and re-encrypt each entry
	for _, vanityPath := range entries {
		realPath := global.GetRealPath(vanityPath)
		encBytes, err := os.ReadFile(realPath)
		if err != nil {
			return errors.New("unable to open \"" + realPath + "\" for decryption: " + err.Error())
		}
		decBytes, err := wrappers.Decrypt(encBytes, oldRCWPassword, false)
		if err != nil {
			return err
		}

		// split & optimize entry
		decSlice := clampTrailingWhitespace(strings.Split(string(decBytes), "\n"))

		// re-encrypt the entry with the new password
		encBytes = wrappers.Encrypt([]byte(strings.Join(decSlice, "\n")), newRCWPassword, true, false)

		// write the entry to the new directory
		if err = os.WriteFile(global.EntryRoot+"-new"+strings.ReplaceAll(vanityPath, "/", global.PathSeparator), encBytes, 0600); err != nil {
			return errors.New("unable to write to file: " + err.Error())
		}
	}

	// swap the new directory with the old one
	if err = os.Rename(global.EntryRoot, global.EntryRoot+"-old"); err != nil {
		return errors.New("unable to rename old directory: " + err.Error())
	}
	if err = os.Rename(global.EntryRoot+"-new", global.EntryRoot); err != nil {
		return errors.New("unable to rename new directory: " + err.Error())
	}

	// generate new sanity check file
	if err = RCWSanityCheckGen(newRCWPassword); err != nil {
		return err
	}

	return nil
}

// VerifyEntries decrypts all entries to memory and returns an error if
// any failures are encountered. Failures likely indicate corrupt entries.
func VerifyEntries(rcwPassword []byte) error {
	entries, _, err := synccommon.WalkEntryDir()
	if err != nil {
		return errors.New("unable to walk entry directory: " + err.Error())
	}
	for _, vanityPath := range entries {
		realPath := global.GetRealPath(vanityPath)
		encBytes, err := os.ReadFile(realPath)
		if err != nil {
			return errors.New("unable to open \"" + realPath + "\" for decryption: " + err.Error())
		}
		decBytes, err := wrappers.Decrypt(encBytes, rcwPassword, false)
		if err != nil {
			return errors.New("unable to verify \"" + vanityPath + "\" (decryption failure): " + err.Error())
		}
		if len(decBytes) < 1 {
			return errors.New("unable to verify \"" + vanityPath + "\" (contains 0/nil bytes)")
		}
	}
	return nil
}

// clampTrailingWhitespace ensures the provided decSlice contains no trailing blank lines.
// If decSlice contains a note, it strips trailing newlines, carriage returns, and tabs from
// each line in the note. Additionally, it removes single trailing spaces and truncates
// multiple trailing spaces to two (for Markdown formatting).
func clampTrailingWhitespace(decSlice []string) []string {
	decSlice = back.RemoveTrailingEmptyStrings(decSlice)
	if len(decSlice) >= 4 {
		for i, noteLine := range decSlice[4:] {
			// remove trailing tabs, carriage returns, and newlines
			noteLine = strings.TrimRight(noteLine, "\t\r\n")

			// determine the number of trailing spaces in the trimmed line
			var endSpacesCount int
			for j := len(noteLine) - 1; j >= 0; j-- {
				if noteLine[j] != ' ' {
					break
				}
				endSpacesCount++
			}

			// remove single spaces, truncate multiple spaces (leave two for Markdown formatting)
			switch endSpacesCount {
			case 0:
				// no trailing spaces
				decSlice[i+4] = noteLine
			case 1:
				// remove the single trailing space
				decSlice[i+4] = strings.TrimRight(noteLine, " ")
			default:
				// truncate the trailing spaces to two
				decSlice[i+4] = noteLine[:len(noteLine)-endSpacesCount+2]
			}
		}
	}
	return decSlice
}

// EntryAddPrecheck ensures the directory meant to contain a new
// entry exists and that realPath is not already used.
// Returns: statusCode (0 = success, 1 = realPath already exists, 2 = containing directory is invalid).
func EntryAddPrecheck(realPath string) (uint8, error) {
	// ensure realPath does not already exist
	isAccessible, _ := back.TargetIsFile(realPath, false) // error is ignored because dir/file status is irrelevant
	if isAccessible {
		return 1, errors.New("target location already exists")
	}
	// ensure target containing directory exists and is not a file
	containingDir := realPath[:strings.LastIndex(realPath, global.PathSeparator)]
	_, err := back.TargetIsFile(containingDir, false)
	if err != nil {
		return 2, errors.New("\"" + containingDir + "\" is not a valid containing directory: " + err.Error())
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

// GenTOTP generates a TOTP token from a secret.
// Prefix `secret` with "steam@" for Steam TOTP format.
func GenTOTP(secret string, time time.Time) (string, error) {
	var totpToken string
	var err error

	if strings.HasPrefix(secret, "steam@") {
		totpToken, err = totp.GenerateCodeCustom(secret[6:], time, totp.ValidateOpts{Period: 30, Digits: 5, Encoder: otp.EncoderSteam})
	} else {
		totpToken, err = totp.GenerateCode(secret, time)
	}

	if err != nil {
		return "", errors.New("unable to generate TOTP token: " + err.Error())
	}

	return totpToken, nil
}
