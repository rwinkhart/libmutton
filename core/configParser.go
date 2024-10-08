package core

import (
	"fmt"
	"io/fs"
	"os"

	"gopkg.in/ini.v1"
)

// loadConfig loads the libmutton.ini file and returns the configuration.
// It is a utility function for ParseConfig and WriteConfig; do not call directly.
func loadConfig() *ini.File {
	cfg, err := ini.Load(ConfigPath)
	if err != nil {
		fmt.Println(AnsiError+"Failed to load libmutton.ini:", err.Error()+AnsiReset)
		os.Exit(ErrorRead)
	}
	return cfg
}

// ParseConfig reads the libmutton.ini file and returns a slice of values for the specified keys.
// Requires: requestedValues (a slice of length 2 arrays each containing a section and a key name),
// missingValueError (an error message to display if a key is missing a value, set to "" for auto-generated or "0" to exit/return silently with code 0).
// Returns: config (slice of values for the specified keys).
func ParseConfig(valuesRequested [][2]string, missingValueError string) []string {
	cfg := loadConfig()

	var config []string

	for _, pair := range valuesRequested {
		value := cfg.Section(pair[0]).Key(pair[1]).String()

		// ensure specified key has a value
		if value == "" {
			switch missingValueError {
			case "":
				fmt.Println(AnsiError + "Failed to find value for key \"" + pair[1] + "\" in section \"[" + pair[0] + "]\" in libmutton.ini" + AnsiReset)
			case "0":
				Exit(0)
			default:
				fmt.Println(AnsiError + missingValueError + AnsiReset)
			}
			os.Exit(ErrorRead)
		}

		config = append(config, value)
	}

	return config
}

// GenDeviceIDList returns a pointer to a slice of all registered device IDs.
// Requires: errorOnFail (set to true to throw an error if the devices directory cannot be read/does not exist)
func GenDeviceIDList(errorOnFail bool) *[]fs.DirEntry {
	// create a slice of all registered devices
	deviceIDList, err := os.ReadDir(ConfigDir + PathSeparator + "devices")
	if err != nil {
		if errorOnFail {
			fmt.Println(AnsiError+"Failed to read the devices directory:", err.Error()+AnsiReset)
			os.Exit(ErrorRead)
		} else {
			return nil // a nil return value indicates that the devices directory could not be read/does not exist
		}
	}
	return &deviceIDList
}

// WriteConfig writes the provided key-value pairs under the specified section headers in the libmutton.ini file.
// Requires: valuesToWrite (a slice of length 3 arrays each containing a section, a key name, and a value).
func WriteConfig(valuesToWrite [][3]string, append bool) {
	var cfg *ini.File

	if append {
		// load existing ini file
		cfg = loadConfig()
	} else {
		// create empty ini container
		cfg = ini.Empty()
	}

	// set all specified key-value pairs in their respective sections
	var section *ini.Section
	for _, trio := range valuesToWrite {
		if cfg.Section(trio[0]) == nil {
			// create and aquire section if it doesn't exist
			section, _ = cfg.NewSection(trio[0])
		} else {
			// acquire existing section
			section = cfg.Section(trio[0])
		}

		// set key-value pair
		section.Key(trio[1]).SetValue(trio[2])
	}

	// save to libmutton.ini
	err := cfg.SaveTo(ConfigPath)
	if err != nil {
		fmt.Println(AnsiError+"Failed to save libmutton.ini:", err.Error()+AnsiReset)
		os.Exit(ErrorWrite)
	}
}
