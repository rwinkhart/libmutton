//go:build darwin

package core

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// copyField copies a field from an entry to the clipboard.
func copyField(executableName, copySubject string) {
	cmd := exec.Command("pbcopy")
	writeToStdin(cmd, copySubject)
	err := cmd.Run()
	if err != nil {
		fmt.Println(AnsiError+"Failed to copy to clipboard:", err.Error()+AnsiReset)
		os.Exit(ErrorClipboard)
	}

	// launch clipboard clearing process if executableName is provided
	if executableName != "" {
		cmd = exec.Command(executableName, "clipclear")
		writeToStdin(cmd, copySubject)
		err = cmd.Start()
		if err != nil {
			fmt.Println(AnsiError + "Failed to launch automated clipboard clearing process - Does this libmutton implementation support the \"clipclear\" argument?" + AnsiReset)
			os.Exit(ErrorClipboard)
		}
		Exit(0) // only exit if clipboard clearing process is launched, otherwise assume continuous clipboard refresh
	}
}

// clipClear is called in a separate process to clear the clipboard after 30 seconds.
func clipClear(oldContents string) {
	time.Sleep(30 * time.Second)

	cmd := exec.Command("pbpaste")
	newContents, err := cmd.Output()
	if err != nil {
		fmt.Println(AnsiError+"Failed to read clipboard contents:", err.Error()+AnsiReset)
		os.Exit(ErrorClipboard)
	}

	if oldContents == strings.TrimRight(string(newContents), "\r\n") {
		cmd = exec.Command("pbcopy")
		writeToStdin(cmd, "")
		err = cmd.Run()
		if err != nil {
			fmt.Println(AnsiError+"Failed to clear clipboard:", err.Error()+AnsiReset)
			os.Exit(ErrorClipboard)
		}
	}
	os.Exit(0) // use os.Exit instead of core.Exit, as this function runs out of a background subprocess that is invisible to the user (will never appear in GUI/TUI environment)
}
