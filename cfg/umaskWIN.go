//go:build windows

package cfg

// setUmask is a dummy function on Windows.
func setUmask(umask int) {
	return
}
