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
