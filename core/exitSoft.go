//go:build returnOnExit

package core

func Exit(code int) {
	return code
}
