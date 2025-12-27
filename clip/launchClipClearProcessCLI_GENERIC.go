//go:build !ios && (!android || termux) && !interactive

package clip

import (
	"os"
	"os/exec"

	"github.com/rwinkhart/go-boilerplate/back"
	"github.com/rwinkhart/libmutton/global"
)

// LaunchClearProcess launches the timed clipboard clearing process.
// For non-interactive CLI implementations, an entirely separate process is created for this purpose.
func LaunchClearProcess(copySubject string) {
	cmd := exec.Command(os.Args[0], "clipclear")
	cmd.SysProcAttr = global.GetSysProcAttr()
	_ = back.WriteToStdin(cmd, copySubject)
	_ = cmd.Start()
	os.Exit(0) // use os.Exit directly since this version of this function is only meant for non-interactive CLI implementations
}
