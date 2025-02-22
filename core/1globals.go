package core

import (
	"os"
)

type ByteInputFetcher func(prompt string) []byte

var (
	PassphraseInputFunction ByteInputFetcher // Clients should set this to a function that fetches hidden input from the user
	Home, _                 = os.UserHomeDir()
)

const (
	LibmuttonVersion = "0.3.0" // Untagged releases feature a letter suffix corresponding to the eventual release version, e.g "0.2.A" -> "0.2.0", "0.2.B" -> "0.2.1"

	FSSpace = "\u259d" // ▝ Space/list separator
	FSPath  = "\u259e" // ▞ Path separator
	FSMisc  = "\u259f" // ▟ Misc. field separator (if \u259d is already used)

	AnsiError = "\033[38;5;9m"
	AnsiReset = "\033[0m"

	ErrorRead             = 101
	ErrorWrite            = 102
	ErrorSyncProcess      = 103
	ErrorServerConnection = 104
	ErrorTargetNotFound   = 105
	ErrorTargetExists     = 106
	ErrorTargetWrongType  = 107
	ErrorDecryption       = 108
	ErrorEncryption       = 109
	ErrorClipboard        = 110
	ErrorOther            = 111
)
