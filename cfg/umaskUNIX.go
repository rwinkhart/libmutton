//go:build !windows

package cfg

import "syscall"

// setUmask sets the file mode creation mask (umask) for the current process.
// The call to syscall.Umask needs to be embedded in another function to allow
// compilation on Windows.
func setUmask(umask int) {
	syscall.Umask(umask)
}
