//go:build !windows && !darwin && !android && !termux && !wsl && !returnOnExit

package core

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

// launchClipClearProcess launches the automated clipboard clearing process.
// For non-interactive CLI implementations, an entirely separate process is created for this purpose.
func launchClipClearProcess(copySubject string, isWayland bool) {
	executableName := os.Args[0]
	cmd := exec.Command(executableName, "clipclear", strconv.FormatBool(isWayland))
	writeToStdin(cmd, copySubject)
	err := cmd.Start()
	if err != nil {
		fmt.Println(AnsiError + "Failed to launch automated clipboard clearing process - Does this libmutton implementation support the \"clipclear\" argument?" + AnsiReset)
		os.Exit(ErrorClipboard)
	}
	os.Exit(0) // use os.Exit directly since this version of this function is only meant for non-interactive CLI implementations
}
