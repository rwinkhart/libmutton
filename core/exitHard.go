//go:build !returnOnExit

package core

import "os"

// Exit (hard) is meant to be used in non-interactive CLI implementations to exit the program after an operation.
func Exit(code int) {
	os.Exit(code)
}
