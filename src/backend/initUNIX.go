//go:build !windows

package backend

import (
	"fmt"
	"os"
)

const FallbackEditor = "vi" // vi is pre-installed on most UNIX systems

// DirInit creates the libmutton directories
func DirInit() {
	// create EntryRoot
	err := os.MkdirAll(EntryRoot, 0700)
	if err != nil {
		fmt.Println(AnsiError + "Failed to create \"" + EntryRoot + "\":" + err.Error() + AnsiReset)
		os.Exit(1)
	}

	// create config directory w/devices subdirectory
	err = os.MkdirAll(ConfigDir+"/devices", 0700)
	if err != nil {
		fmt.Println(AnsiError + "Failed to create \"" + ConfigDir + "\":" + err.Error() + AnsiReset)
		os.Exit(1)
	}
}

// textEditorFallback returns the value of the $EDITOR environment variable, or FallbackEditor if it is not set
func textEditorFallback() string {
	// ensure textEditor is set
	textEditor := os.Getenv("EDITOR")
	if textEditor == "" {
		textEditor = FallbackEditor
	}
	return textEditor
}
