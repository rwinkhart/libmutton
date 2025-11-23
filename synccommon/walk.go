package synccommon

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/rwinkhart/libmutton/global"
)

// WalkEntryDir walks the entry directory and returns lists of all files and
// directories found (two separate lists) relative to the libmutton entry root.
// Regardless of platform, all paths are stored with forward slashes (UNIX-style).
func WalkEntryDir() ([]string, []string, error) {
	// define file/directory containing slices so that they may be accessed by the anonymous WalkDir function
	var fileList, dirList []string

	// walk entry directory
	err := filepath.WalkDir(global.EntryRoot,
		func(realPath string, entry fs.DirEntry, err error) error {

			// check for errors encountered while walking directory
			if err != nil {
				if os.IsNotExist(err) {
					return errors.New("entry directory does not exist; initialize libmutton to create it")
				} else {
					return errors.New("an unexpected error occurred while generating the entry list: " + err.Error())
				}
			}

			// trim root path from each path before storing and replace backslashes with forward slashes
			vanityPath := global.GetVanityPath(realPath)

			// append the path to the appropriate slice
			if !entry.IsDir() {
				fileList = append(fileList, vanityPath)
			} else {
				dirList = append(dirList, vanityPath)
			}

			return nil
		})
	if err != nil {
		return nil, nil, err
	}
	return fileList, dirList, nil
}
