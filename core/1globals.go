package core

import (
	"os"
)

var (
	Home, _ = os.UserHomeDir()
)

const (
	AnsiError        = "\033[38;5;9m"
	AnsiReset        = "\033[0m"
	LibmuttonVersion = "0.2.B" // untagged releases feature a letter suffix corresponding to the eventual release version, e.g "0.2.A" -> "0.2.0", "0.2.B" -> "0.2.1"
)

// numeric error codes legend
// 0: no error
// 101: read error
// 102: write error
// 103: sync process error
// 104: server connection error
// 105: target not found
// 106: target already exists
// 107: target wrong type (file/directory)
// 108: decryption error
// 109: encryption error
// 110: clipboard error
// 111: other error
