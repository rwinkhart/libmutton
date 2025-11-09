package global

type ByteInputFetcher func(prompt string) []byte

var (
	GetPassword ByteInputFetcher // Clients should set this to a function that fetches hidden input from the user
)

const (
	LibmuttonVersion = "0.4.2" // Untagged releases feature a letter suffix corresponding to the eventual release version, e.g "0.2.A" -> "0.2.0", "0.2.B" -> "0.2.1"

	FSSpace = "\u259d" // ▝ Space/list separator
	FSPath  = "\u259e" // ▞ Path separator
	FSMisc  = "\u259f" // ▟ Misc. field separator (if \u259d is already used)

	ErrorSyncProcess = 104
	ErrorDecryption  = 105
	ErrorEncryption  = 106
	ErrorClipboard   = 107
)
