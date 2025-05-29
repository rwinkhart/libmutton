//go:build windows

package core

// setUmask is a dummy function on Windows.
func setUmask(umask int) {
	return
}
