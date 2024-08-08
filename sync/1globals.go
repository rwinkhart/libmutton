package sync

import "github.com/rwinkhart/libmutton/core"

// define field separator constants
const (
	FSSpace = "\u259d" // ▝ space/list separator
	FSPath  = "\u259e" // ▞ path separator
	FSMisc  = "\u259f" // ▟ misc. field separator (if \u259d is already used)
)

// rootLength stores length of core.EntryRoot string
var rootLength = len(core.EntryRoot)
