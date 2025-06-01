package core

import (
	"errors"
	"os"
	"strings"

	"github.com/rwinkhart/go-boilerplate/back"
	"github.com/rwinkhart/libmutton/crypt"
	"github.com/rwinkhart/libmutton/global"
	"github.com/rwinkhart/libmutton/synccommon"
	"github.com/rwinkhart/rcw/wrappers"
)

// WriteEntry writes entryData to an encrypted file at targetLocation.
func WriteEntry(targetLocation string, decBytes []byte) error {
	err := os.WriteFile(targetLocation, crypt.EncryptBytes(decBytes), 0600)
	if err != nil {
		return errors.New("unable to write to file: " + err.Error())
	}
	return nil
}

// EntryRefresh re-encrypts all libmutton entries with a new passphrase
// and optimizes each entry to ensure they are as slim as possible.
// This includes stripping trailing whitespace/newlines/carriage returns
// from each field and running each note through ClampTrailingWhitespace
// to ensure each note line is optimized as possible without breaking
// Markdown formatting.
// Be sure to verify passphrases before using as input for this function!!
func EntryRefresh(oldRCWPassphrase, newRCWPassphrase []byte, removeOldDir bool) error {
	// ensure global.EntryRoot+"-new" and global.EntryRoot-"old" do not exist
	dirEnds := []string{"-new", "-old"}
	for i, dirEnd := range dirEnds {
		if i == 1 && !removeOldDir {
			if _, err := os.Stat(global.EntryRoot + "-old"); !os.IsNotExist(err) {
				return errors.New("unable to refresh entries: \"" + global.EntryRoot + "-old\" already exists")
			}
		}
		err := os.RemoveAll(global.EntryRoot + dirEnd)
		if err != nil {
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
		err := os.MkdirAll(fullPath, 0700)
		if err != nil {
			return errors.New("unable to create temporary directory \"" + fullPath + "\": " + err.Error())
		}
	}

	// decrypt, optimize, and re-encrypt each entry
	for _, entryName := range entries {
		targetLocation := global.TargetLocationFormat(entryName)
		encBytes, err := os.ReadFile(targetLocation)
		if err != nil {
			return errors.New("unable to open \"" + targetLocation + "\" for decryption: " + err.Error())
		}
		decBytes, err := wrappers.Decrypt(encBytes, oldRCWPassphrase)
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

		// re-encrypt the entry with the new passphrase
		encBytes = wrappers.Encrypt([]byte(strings.Join(decryptedEntry, "\n")), newRCWPassphrase)

		// write the entry to the new directory
		err = os.WriteFile(global.EntryRoot+"-new"+strings.ReplaceAll(entryName, "/", global.PathSeparator), encBytes, 0600)
		if err != nil {
			return errors.New("unable to write to file: " + err.Error())
		}

		// generate new sanity check file
		err = RCWSanityCheckGen(newRCWPassphrase)
		if err != nil {
			return err
		}
	}

	// swap the new directory with the old one
	err = os.Rename(global.EntryRoot, global.EntryRoot+"-old")
	if err != nil {
		return errors.New("unable to rename old directory: " + err.Error())
	}
	err = os.Rename(global.EntryRoot+"-new", global.EntryRoot)
	if err != nil {
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
// entry exists and that the target entry location is not already used.
// Returns: statusCode (0 = success, 1 = target location already exists, 2 = containing directory is invalid).
func EntryAddPrecheck(targetLocation string) (uint8, error) {
	// ensure target location does not already exist
	_, isAccessible, _ := back.TargetIsFile(targetLocation, false, 0) // error is ignored because errorOnFail is false
	if isAccessible {
		return 1, errors.New("target location already exists")
	}
	// ensure target containing directory exists and is a directory (not a file)
	containingDir := targetLocation[:strings.LastIndex(targetLocation, global.PathSeparator)]
	isFile, isAccessible, _ := back.TargetIsFile(containingDir, false, 1) // error is ignored because errorOnFail is false
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
