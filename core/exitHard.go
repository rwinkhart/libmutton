//go:build !returnOnExit

package core

import "os"

func Exit(code int) {
	os.Exit(code)
}
