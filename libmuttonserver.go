package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rwinkhart/libmutton/core"
	"github.com/rwinkhart/libmutton/sync"
)

const ansiBold = "\033[1m"

func main() {
	args := os.Args
	if len(args) < 2 {
		helpServer()
	}

	// check if stdin was provided
	stdinInfo, _ := os.Stdin.Stat()
	stdinPresent := stdinInfo.Mode()&os.ModeNamedPipe != 0

	var stdin []string
	if stdinPresent {
		// read stdin, appending each line to the stdin slice
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			stdin = append(stdin, scanner.Text())
		}
	}

	switch args[1] {
	case "fetch":
		// print all information needed for syncing to stdout for interpretation by the client
		// stdin[0] is expected to be the device ID
		sync.GetRemoteDataFromServer(stdin[0])
	case "rename":
		// move an entry to a new location before using fallthrough to add its previous iteration to the deletions directory
		// stdin[0] is evaluated after fallthrough
		// stdin[1] is expected to be the OLD incomplete target location with FSPath representing path separators - Always pass in UNIX format
		// stdin[2] is expected to be the NEW incomplete target location with FSPath representing path separators - Always pass in UNIX format
		sync.RenameLocal(strings.ReplaceAll(stdin[1], core.FSPath, "/"), strings.ReplaceAll(stdin[2], core.FSPath, "/"), true)
		fallthrough // fallthrough to add the old entry to the deletions directory
	case "shear":
		// shear an entry from the server and add it to the deletions directory
		// stdin[0] is expected to be the device ID
		// stdin[1] is expected to be the incomplete target location with FSPath representing path separators - Always pass in UNIX format
		sync.ShearLocal(strings.ReplaceAll(stdin[1], core.FSPath, "/"), stdin[0])
	case "addfolder":
		// add a new folder to the server
		// stdin[0] is expected to be the incomplete target location with FSPath representing path separators - Always pass in UNIX format
		sync.AddFolderLocal(strings.ReplaceAll(stdin[0], core.FSPath, "/"))
	case "register":
		// register a new device ID
		// stdin[0] is expected to be the device ID
		// stdin[1] is expected to be the old device ID (for removal)
		fileToClose, _ := os.OpenFile(core.ConfigDir+core.PathSeparator+"devices"+core.PathSeparator+stdin[0], os.O_CREATE|os.O_WRONLY, 0600) // errors ignored; failure unlikely to occur if init was successful; "register" is not a user-facing argument and thus the error would not be visible
		_ = fileToClose.Close()
		if stdin[1] != core.FSMisc { // sync.FSMisc is used to indicate that no device ID is being replaced
			_ = os.RemoveAll(core.ConfigDir + core.PathSeparator + "devices" + core.PathSeparator + stdin[1])
		}
		// print EntryRoot and bool indicating OS type to stdout for client to store in config
		fmt.Print(core.EntryRoot + core.FSSpace + strconv.FormatBool(core.IsWindows))
	case "init":
		// create the necessary directories for libmuttonserver to function
		core.DirInit(false)
		_ = os.MkdirAll(core.ConfigDir+core.PathSeparator+"deletions", 0700) // error ignored; failure would have occurred by this point in core.DirInit
		fmt.Println("libmuttonserver directories initialized")
	case "version":
		versionServer()
	default:
		helpServer()
	}
}

func helpServer() {
	fmt.Print(ansiBold + "\nlibmuttonserver | Copyright (c) 2024 Randall Winkhart\n" + core.AnsiReset + `
This software exists under the MIT license; you may redistribute it under certain conditions.
This program comes with absolutely no warranty; type "libmuttonserver version" for details.

` + ansiBold + "Usage:" + core.AnsiReset + ` libmuttonserver <argument>

` + ansiBold + "Arguments (user):" + core.AnsiReset + `
 help                    Bring up this menu
 version                 Display version and license information
 init                    Create the necessary directories for libmuttonserver to function` + "\n\n")
	os.Exit(0)
}

func versionServer() {
	fmt.Print(ansiBold + "\n                    MIT License" + core.AnsiReset + `

  Permission is hereby granted, free of charge, to any
person obtaining a copy of this software and associated
  documentation files (the "Software"), to deal in the
    Software without restriction, including without
   limitation the rights to use, copy, modify, merge,
 publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software
    is furnished to do so, subject to the following
                      conditions:

 The above copyright notice and this permission notice
shall be included in all copies or substantial portions
                   of the Software.

 THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF
ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
  TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
  PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT
 SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR
 ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN
 ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE
           OR OTHER DEALINGS IN THE SOFTWARE.` + "\n\n---------------------------------------------------------")
	fmt.Print(ansiBold + "\n\n              libmuttonserver" + core.AnsiReset + " Version " + core.LibmuttonVersion + `

           Copyright (c) 2024 Randall Winkhart` + "\n\n")
}
