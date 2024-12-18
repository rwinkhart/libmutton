//go:build !windows && !darwin && !termux && !wsl

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
	var envSet bool // track whether environment variables are set
	var cmd *exec.Cmd
	// determine whether to use wl-copy (Wayland) or xclip (X11)
	if _, envSet = os.LookupEnv("WAYLAND_DISPLAY"); envSet {
		cmd = exec.Command("wl-copy", "-t", "text/plain")
	} else if _, envSet = os.LookupEnv("DISPLAY"); envSet {
		cmd = exec.Command("xclip", "-sel", "c", "-t", "text/plain")
	} else {
		fmt.Println(AnsiError + "Clipboard platform could not be determined - Note that the clipboard does not function in a raw TTY" + AnsiReset)
		os.Exit(ErrorClipboard)
	}

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

	// determine clipboard tool to use (wl-clipboard VS xclip)
	var envSet bool
	var cmdClear, cmdPaste *exec.Cmd
	if _, envSet = os.LookupEnv("WAYLAND_DISPLAY"); envSet {
		cmdClear = exec.Command("wl-copy", "-c")
		cmdPaste = exec.Command("wl-paste")
	} else if _, envSet = os.LookupEnv("DISPLAY"); envSet {
		cmdClear = exec.Command("xclip", "-i", "/dev/null", "-sel", "c")
		cmdPaste = exec.Command("xclip", "-o", "-sel", "c")
	} else {
		fmt.Println(AnsiError + "Clipboard platform could not be determined - Neither $WAYLAND_DISPLAY nor $DISPLAY are set" + AnsiReset)
		os.Exit(ErrorClipboard)
	}

	// read current clipboard contents
	newContents, err := cmdPaste.Output()
	if err != nil {
		fmt.Println(AnsiError+"Failed to read clipboard contents:", err.Error()+AnsiReset)
		os.Exit(ErrorClipboard)
	}

	// clear clipboard if contents have not been modified
	if oldContents == strings.TrimRight(string(newContents), "\r\n") {
		err = cmdClear.Run()
		if err != nil {
			fmt.Println(AnsiError+"Failed to clear clipboard:", err.Error()+AnsiReset)
			os.Exit(ErrorClipboard)
		}
	}
	os.Exit(0) // use os.Exit instead of core.Exit, as this function runs out of a background subprocess that is invisible to the user (will never appear in GUI/TUI environment)
}
