package config

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"

	"github.com/rwinkhart/libmutton/global"
)

type CfgT struct {
	Libmutton struct {
		OfflineMode      *bool   `json:"offlineMode"`
		SSHUser          *string `json:"sshUser"`
		SSHIP            *string `json:"sshIP"`
		SSHPort          *string `json:"sshPort"`
		SSHEntryRootPath *string `json:"sshEntryRootPath"`
		SSHAgeDirPath    *string `json:"sshAgeDirPath"`
		SSHKeyPath       *string `json:"sshKeyPath"`
		SSHKeyProtected  *bool   `json:"sshKeyProtected"`
		SSHIsWindows     *bool   `json:"sshIsWindows"`
	} `json:"libmutton"`
	ThirdParty *map[string]any `json:"thirdParty"`
}

// Load loads libmuttoncfg.json and returns the configuration.
func Load() (*CfgT, error) {
	cfgBytes, err := os.ReadFile(global.CfgPath)
	if err != nil {
		return nil, errors.New("unable to load libmuttoncfg.json: " + err.Error())
	}
	var cfg CfgT
	err = json.Unmarshal(cfgBytes, &cfg)
	if err != nil {
		return nil, errors.New("unable to unmarshal libmuttoncfg.json: " + err.Error())
	}
	return &cfg, nil
}

// Write writes cfg to libmuttoncfg.json.
// If used in append mode, any nil values in the
// input cfg will be substituted with the existing values.
func Write(cfg *CfgT, appendMode bool) error {
start:
	if appendMode {
		// check if any fields are nil
		var hasNilFields bool
		cfgValue := reflect.ValueOf(&cfg.Libmutton).Elem()
		for i := 0; i < cfgValue.NumField(); i++ {
			field := cfgValue.Field(i)
			if field.IsNil() {
				hasNilFields = true
				break
			}
		}

		// load old cfg and copy nil fields
		if hasNilFields {
			oldCfg, err := Load()
			if err != nil {
				// failed to load config, leave append mode
				appendMode = false
				goto start
			}
			oldValue := reflect.ValueOf(&oldCfg.Libmutton).Elem()
			for i := 0; i < cfgValue.NumField(); i++ {
				field := cfgValue.Field(i)
				if field.IsNil() {
					field.Set(oldValue.Field(i))
				}
			}
		}
	}
	cfgBytes, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return errors.New("unable to marshal new/updated cfg: " + err.Error())
	}
	err = os.WriteFile(global.CfgPath, cfgBytes, 0600)
	if err != nil {
		return errors.New("unable to write new/updated cfg to libmuttoncfg.json: " + err.Error())
	}

	return nil
}
