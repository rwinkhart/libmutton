package core

type ByteInputFetcher func(prompt string) []byte

var (
	GetPassphrase ByteInputFetcher // Clients should set this to a function that fetches hidden input from the user
)

const (
	LibmuttonVersion = "0.E.0" // Untagged releases feature a letter suffix corresponding to the eventual release version, e.g "0.2.A" -> "0.2.0", "0.2.B" -> "0.2.1"

	FSSpace = "\u259d" // ▝ Space/list separator
	FSPath  = "\u259e" // ▞ Path separator
	FSMisc  = "\u259f" // ▟ Misc. field separator (if \u259d is already used)

	ErrorSyncProcess      = 103
	ErrorServerConnection = 104
	ErrorTargetExists     = 106
	ErrorDecryption       = 108
	ErrorEncryption       = 109
	ErrorClipboard        = 110
)
