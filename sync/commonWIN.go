//go:build windows

package sync

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/rwinkhart/libmutton/core"
)

// WalkEntryDir walks the entry directory and returns lists of all files and directories found (two separate lists).
// Regardless of platform, all paths are stored with forward slashes (UNIX-style).
func WalkEntryDir() ([]string, []string) {
	// define file/directory containing slices so that they may be accessed by the anonymous WalkDir function
	var fileList []string
	var dirList []string

	// walk entry directory
	_ = filepath.WalkDir(core.EntryRoot,
		func(fullPath string, entry fs.DirEntry, err error) error {

			// check for errors encountered while walking directory
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Println(core.AnsiError + "The entry directory does not exist - Initialize libmutton to create it" + core.AnsiReset)
				} else {
					// otherwise, print the source of the error
					fmt.Println(core.AnsiError+"An unexpected error occurred while generating the entry list:", err.Error()+core.AnsiReset)
				}
				os.Exit(core.ErrorOther)
			}

			// trim root path from each path before storing and replace backslashes with forward slashes
			trimmedPath := strings.ReplaceAll(fullPath[rootLength:], "\\", "/")

			// append the path to the appropriate slice
			if !entry.IsDir() {
				fileList = append(fileList, trimmedPath)
			} else {
				dirList = append(dirList, trimmedPath)
			}

			return nil
		})

	return fileList, dirList
}
