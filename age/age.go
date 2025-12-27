package age

import (
	"crypto/rand"
	"errors"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/rwinkhart/go-boilerplate/back"
	"github.com/rwinkhart/libmutton/crypt"
	"github.com/rwinkhart/libmutton/global"
	"github.com/rwinkhart/libmutton/synccommon"
)

// Entry creates updates the age file for a vanity path.
func Entry(vanityPath string, timestamp int64) error {
	ageFilePath := global.AgeDir + global.PathSeparator + strings.ReplaceAll(vanityPath, "/", global.FSPath)
	f, err := os.OpenFile(ageFilePath, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return errors.New("unable to create age file for " + vanityPath + ": " + err.Error())
	}
	_ = f.Close() // error ignored; if the file could be created, it can probably be closed
	err = os.Chtimes(ageFilePath, time.Now(), time.Unix(timestamp, 0))
	if err != nil {
		return errors.New("unable to set timestamp on age file for " + vanityPath + ": " + err.Error())
	}
	return nil
}

// AllPasswordEntries adds age data for all un-aged entries containing passwords.
// Each entry is aged with a random timestamp from within the last year to prevent
// all entries having their passwords expire at the same time.
func AllPasswordEntries(forceReage bool) error {
	allVanityPaths, _, err := synccommon.WalkEntryDir()
	if err != nil {
		return errors.New("unable to walk entry directory: " + err.Error())
	}

	for _, vanityPath := range allVanityPaths {
		// ensure entry is not already aged (unless forcing re-age)
		if !forceReage {
			// ignore error; we only care if we can access the path or not
			isAccessible, _ := back.TargetIsFile(global.AgeDir+global.PathSeparator+strings.ReplaceAll(vanityPath, "/", global.FSPath), true)
			if isAccessible {
				// entry already aged, skip it
				continue
			}
		}

		decSlice, err := crypt.DecryptFileToSlice(global.EntryRoot + vanityPath)
		if err != nil {
			return err
		}
		if decSlice != nil && decSlice[0] != "" {
			// calculate random UNIX timestamp from within the last 365 days
			offsetInt, _ := rand.Int(rand.Reader, big.NewInt(31557600))
			randomOffset := time.Duration(offsetInt.Int64()) * time.Second
			err = Entry(vanityPath, time.Now().Add(-randomOffset).Unix())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// TranslateAgeTimestamp returns a uint8 value indicating
// the interpreted age status of an age timestamp associated
// with an entry.
// Magic number legend:
// 0 -> no age, 1 -> fresh, 2 -> expiring soon (within a month), 3 -> expired
func TranslateAgeTimestamp(timestamp int64) uint8 {
	if timestamp == 0 {
		return 0
	}
	daysOld := time.Since(time.Unix(timestamp, 0)).Hours() / 24
	if daysOld >= 365 {
		return 3 // expired
	} else if daysOld >= 335 {
		return 2 // expiring soon
	}
	return 1 // fresh
}
