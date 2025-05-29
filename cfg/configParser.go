package cfg

import (
	"fmt"

	"github.com/rwinkhart/go-boilerplate/back"
	"github.com/rwinkhart/libmutton/global"
	"gopkg.in/ini.v1"
)

// loadConfig loads the libmutton.ini file and returns the configuration.
// It is a utility function for ParseConfig and WriteConfig; do not call directly.
func loadConfig() *ini.File {
	cfg, err := ini.Load(global.ConfigPath)
	if err != nil {
		back.PrintError("Failed to load libmutton.ini: "+err.Error(), back.ErrorRead, true)
	}
	return cfg
}

// ParseConfig reads the libmutton.ini file and returns a slice of values for the specified keys.
// Requires: valuesRequested (a slice of length 2 arrays each containing a section and a key name),
// missingValueError (an error message to display if a key is missing a value, set to "" for auto-generated or "0" to exit/return silently with code 0).
// Returns: config (slice of values for the specified keys),
// error (nil if no error occurred, otherwise an error using the generated or provided message).
func ParseConfig(valuesRequested [][2]string, missingValueError string) ([]string, error) {
	var err error
	cfg := loadConfig()

	var config []string

	for _, pair := range valuesRequested {
		value := cfg.Section(pair[0]).Key(pair[1]).String()

		// ensure specified key has a value
		if value == "" {
			switch missingValueError {
			case "":
				err = fmt.Errorf("failed to find value for key \"%s\" in section \"[%s]\" in libmutton.ini", pair[1], pair[0])
			case "0":
				back.Exit(0) // hard (expected) exit for CLI; GUI/TUI continue silently
			default:
				err = fmt.Errorf("%s", missingValueError)
			}
			back.PrintError(err.Error(), back.ErrorRead, false)
			// if interactive (soft exit), return nil and the error to be handled by the caller
			return nil, err
		}

		config = append(config, value)
	}

	return config, err
}

// WriteConfig writes the provided key-value pairs under the specified section headers in the libmutton.ini file.
// Requires: valuesToWrite (a slice of length 3 arrays each containing a section, a key name, and a value),
// prune (a slice similar to valuesToWrite to allow removing the specified keys from an existing config),
// append (set to true to append to the existing libmutton.ini file, false to overwrite it).
func WriteConfig(valuesToWrite [][3]string, keysToPrune [][2]string, append bool) {
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
			// create and acquire section if it doesn't exist
			section, _ = cfg.NewSection(trio[0])
		} else {
			// acquire existing section
			section = cfg.Section(trio[0])
		}

		// set key-value pair
		section.Key(trio[1]).SetValue(trio[2])
	}

	// prune specified keys from the existing config
	if append && len(keysToPrune) > 0 {
		// remove specified keys pairs from the existing config
		for _, pair := range keysToPrune {
			cfg.Section(pair[0]).DeleteKey(pair[1])
		}
	}

	// save to libmutton.ini
	setUmask(0077) // only give permissions to owner
	err := cfg.SaveTo(global.ConfigPath)
	if err != nil {
		back.PrintError("Failed to save libmutton.ini: "+err.Error(), back.ErrorWrite, true)
	}
}
