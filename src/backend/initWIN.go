//go:build windows

package backend

import (
	"fmt"
	"os"
)

const FallbackEditor = "nvim" // since there is no pre-installed CLI editor on Windows, default to the most popular one

// DirInit creates the libmutton directories
func DirInit() {
	// create EntryRoot (includes config directory on Windows)
	err := os.MkdirAll(EntryRoot, 0700)
	if err != nil {
		fmt.Println(AnsiError + "Failed to create \"" + EntryRoot + "\":" + err.Error() + AnsiReset)
		os.Exit(1)
	}
	os.MkdirAll(ConfigDir+"\\devices", 0700)
}

// textEditorFallback returns FallbackEditor
func textEditorFallback() string {
	return FallbackEditor
}
