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

// WriteEntry writes entryData to an encrypted file at realPath.
// If the entry contains an updated password, an age file is also created.
func WriteEntry(realPath string, decSlice []string, passwordIsNew bool) error {
	err := os.WriteFile(realPath, crypt.EncryptBytes([]byte(strings.Join(decSlice, "\n"))), 0600)
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
// This includes stripping trailing whitespace/newlines/carriage returns
// from each field and running each note through ClampTrailingWhitespace
// to ensure each note line is optimized as possible without breaking
// Markdown formatting.
// Be sure to verify passwords before using as input for this function!!
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
		decBytes, err := wrappers.Decrypt(encBytes, oldRCWPassword)
		decryptedEntry := strings.Split(string(decBytes), "\n")
		if err != nil {
			return err
		}
		// strip trailing whitespace...
		fieldsLength := len(decryptedEntry)
		if fieldsLength < 4 {
			fieldsMain := back.RemoveTrailingEmptyStrings(decryptedEntry)
			// ...from each non-note field
			for i, line := range fieldsMain {
				fieldsMain[i] = strings.TrimRight(line, " \t\r\n")
			}
			decryptedEntry = fieldsMain
		} else {
			fieldsMain := decryptedEntry[:4]
			fieldsNote := back.RemoveTrailingEmptyStrings(decryptedEntry[4:])
			// ...from each non-note field
			for i, line := range fieldsMain {
				fieldsMain[i] = strings.TrimRight(line, " \t\r\n")
			}
			// ...and from each note line (preserve Markdown formatting)
			ClampTrailingWhitespace(fieldsNote)

			// re-combine fields
			decryptedEntry = append(fieldsMain, fieldsNote...)
		}

		// re-encrypt the entry with the new password
		encBytes = wrappers.Encrypt([]byte(strings.Join(decryptedEntry, "\n")), newRCWPassword)

		// write the entry to the new directory
		if err = os.WriteFile(global.EntryRoot+"-new"+strings.ReplaceAll(vanityPath, "/", global.PathSeparator), encBytes, 0600); err != nil {
			return errors.New("unable to write to file: " + err.Error())
		}

		// generate new sanity check file
		if err = RCWSanityCheckGen(newRCWPassword); err != nil {
			return err
		}
	}

	// swap the new directory with the old one
	if err = os.Rename(global.EntryRoot, global.EntryRoot+"-old"); err != nil {
		return errors.New("unable to rename old directory: " + err.Error())
	}
	if err = os.Rename(global.EntryRoot+"-new", global.EntryRoot); err != nil {
		return errors.New("unable to rename new directory: " + err.Error())
	}

	return nil
}

// ClampTrailingWhitespace strips trailing newlines, carriage returns, and tabs from each line in a note.
// Additionally, it removes single trailing spaces and truncates multiple trailing spaces to two (for Markdown formatting).
func ClampTrailingWhitespace(note []string) {
	for i, line := range note {
		// remove trailing tabs, carriage returns, and newlines
		line = strings.TrimRight(line, "\t\r\n")

		// determine the number of trailing spaces in the trimmed line
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
			// no trailing spaces
			note[i] = line
		case 1:
			// remove the single trailing space
			note[i] = strings.TrimRight(line, " ")
		default:
			// truncate the trailing spaces to two
			note[i] = line[:len(line)-endSpacesCount+2]
		}
	}
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
