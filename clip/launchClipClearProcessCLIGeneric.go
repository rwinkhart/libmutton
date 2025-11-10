//go:build (windows || darwin || android || ios || termux || wsl) && !interactive

package clip

import (
	"os"
	"os/exec"

	"github.com/rwinkhart/go-boilerplate/back"
)

// LaunchClipClearProcess launches the timed clipboard clearing process.
// For non-interactive CLI implementations, an entirely separate process is created for this purpose.
func LaunchClipClearProcess(copySubject string) {
	cmd := exec.Command(os.Args[0], "clipclear")
	_ = back.WriteToStdin(cmd, copySubject)
	_ = cmd.Start()
	os.Exit(0) // use os.Exit directly since this version of this function is only meant for non-interactive CLI implementations
}
