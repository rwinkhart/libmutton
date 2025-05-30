package cfg

import (
	"errors"
	"fmt"

	"github.com/rwinkhart/libmutton/global"
	"gopkg.in/ini.v1"
)

// loadConfig loads the libmutton.ini file and returns the configuration.
// It is a utility function for ParseConfig and WriteConfig; do not call directly.
func loadConfig() (*ini.File, error) {
	cfg, err := ini.Load(global.ConfigPath)
	if err != nil {
		return nil, errors.New("unable to load libmutton.ini: " + err.Error())
	}
	return cfg, nil
}

// ParseConfig reads the libmutton.ini file and returns a slice of values for the specified keys.
// Requires: valuesRequested (a slice of length 2 arrays each containing a section and a key name).
// Returns: config (slice of values for the specified keys).
// If requesting SSH config, request "LIBMUTTON/offlineMode" first to avoid errors.
func ParseConfig(valuesRequested [][2]string) ([]string, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}

	var config []string
	for _, pair := range valuesRequested {
		value := cfg.Section(pair[0]).Key(pair[1]).String()
		// ensure specified key has a value
		if value == "" {
			return nil, fmt.Errorf("unable to find value for key \"%s\" in section \"[%s]\" in libmutton.ini", pair[1], pair[0])
		}
		config = append(config, value)

		// notify requester immediately if in offline mode
		if pair[0] == "LIBMUTTON" && pair[1] == "offlineMode" && value == "true" {
			return config, nil
		}
	}

	return config, err
}

// WriteConfig writes the provided key-value pairs under the specified section headers in the libmutton.ini file.
// Requires: valuesToWrite (a slice of length 3 arrays each containing a section, a key name, and a value),
// prune (a slice similar to valuesToWrite to allow removing the specified keys from an existing config),
// append (set to true to append to the existing libmutton.ini file, false to overwrite it).
func WriteConfig(valuesToWrite [][3]string, keysToPrune [][2]string, append bool) error {
	var cfg *ini.File
	var err error

	if append {
		// load existing ini file
		cfg, err = loadConfig()
		if err != nil {
			return errors.New("unable to load existing libmutton.ini: " + err.Error())
		}
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
	err = cfg.SaveTo(global.ConfigPath)
	if err != nil {
		return errors.New("unable to save libmutton.ini: " + err.Error())
	}
	return nil
}
