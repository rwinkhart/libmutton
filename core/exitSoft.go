//go:build interactive

package core

// Exit (soft) is meant to be used in interactive implementations (GUIs/TUIs) to keep the program running after an operation.
func Exit(code int) int {
	return code
}
