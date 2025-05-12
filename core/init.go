package core

import (
	"github.com/rwinkhart/go-boilerplate/back"
	"github.com/rwinkhart/libmutton/global"
	"github.com/rwinkhart/rcw/wrappers"
)

// RCWSanityCheckGen generates the RCW sanity check file for libmutton.
func RCWSanityCheckGen(passphrase []byte) {
	err := wrappers.GenSanityCheck(global.ConfigDir+global.PathSeparator+"sanity.rcw", passphrase)
	if err != nil {
		back.PrintError("Failed to generate sanity check file: "+err.Error(), back.ErrorWrite, true)
	}
}
