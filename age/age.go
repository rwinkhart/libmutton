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

// Entry creates/updates the age file for a vanity path.
// It does not touch the actual entry, meaning it won't be synced;
// this is because this function is meant to be used on entries that
// are being created or modified, so they will sync. If using this
// to age an entry without modifying the entry, the caller must also
// update the mod time on the entry to trigger a sync.
func Entry(vanityPath string, timestamp int64) error {
	ageFilePath := global.AgeDir + global.PathSeparator + strings.ReplaceAll(vanityPath, "/", global.FSPath)
	f, err := os.OpenFile(ageFilePath, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return errors.New("unable to create age file for " + vanityPath + ": " + err.Error())
	}
	_ = f.Close() // error ignored; if the file could be created, it can probably be closed
	if err = os.Chtimes(ageFilePath, time.Now(), time.Unix(timestamp, 0)); err != nil {
		return errors.New("unable to set timestamp on age file for " + vanityPath + ": " + err.Error())
	}
	return nil
}

// AllPasswordEntries adds age data for all un-aged entries containing passwords.
// Each entry is aged with a random timestamp from within the last year to prevent
// all entries having their passwords expire at the same time.
// It also updates the mod time on the actual entry to trigger a sync.
// Leave rcwPassword nil to use RCW demonization.
func AllPasswordEntries(forceReage bool, rcwPassword []byte) error {
	allVanityPaths, _, err := synccommon.WalkEntryDir()
	if err != nil {
		return errors.New("unable to walk entry directory: " + err.Error())
	}
	now := time.Now()

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

		decSlice, err := crypt.DecryptFileToSlice(global.EntryRoot+vanityPath, rcwPassword)
		if err != nil {
			return err
		}
		if decSlice != nil && decSlice[0] != "" {
			// calculate random UNIX timestamp from within the last 365 days
			offsetInt, _ := rand.Int(rand.Reader, big.NewInt(31557600))
			randomOffset := time.Duration(offsetInt.Int64()) * time.Second
			if err = Entry(vanityPath, now.Add(-randomOffset).Unix()); err != nil {
				return err
			}
			// update entry mod time to trigger sync
			if err = os.Chtimes(global.GetRealPath(vanityPath), now, now); err != nil {
				return errors.New("unable to update mod time on entry (" + vanityPath + "): " + err.Error())
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
func TranslateAgeTimestamp(timestamp *int64) uint8 {
	if timestamp == nil {
		return 0
	}
	daysOld := time.Since(time.Unix(*timestamp, 0)).Hours() / 24
	if daysOld >= 365 {
		return 3 // expired
	} else if daysOld >= 335 {
		return 2 // expiring soon
	}
	return 1 // fresh
}
