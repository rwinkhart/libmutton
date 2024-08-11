package sync

import "github.com/rwinkhart/libmutton/core"

const (
	FSSpace = "\u259d" // ▝ space/list separator
	FSPath  = "\u259e" // ▞ path separator
	FSMisc  = "\u259f" // ▟ misc. field separator (if \u259d is already used)
)

var rootLength = len(core.EntryRoot) // length of core.EntryRoot string
