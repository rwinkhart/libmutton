//go:build darwin

package core

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// copyString copies a string to the clipboard.
func copyString(continuous bool, copySubject string) {
	cmd := exec.Command("pbcopy")
	writeToStdin(cmd, copySubject)
	err := cmd.Run()
	if err != nil {
		fmt.Println(AnsiError+"Failed to copy to clipboard:", err.Error()+AnsiReset)
		os.Exit(ErrorClipboard)
	}

	if !continuous {
		launchClipClear(copySubject)
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
	Exit(0)
}
