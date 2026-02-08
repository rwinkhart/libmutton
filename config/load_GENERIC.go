//go:build !ios

package config

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/rwinkhart/libmutton/global"
)

// Load loads libmuttoncfg.json and returns the configuration.
func Load() (*CfgT, error) {
	cfgBytes, err := os.ReadFile(global.CfgPath)
	if err != nil {
		return nil, errors.New("unable to load libmuttoncfg.json: " + err.Error())
	}
	var cfg CfgT
	if err = json.Unmarshal(cfgBytes, &cfg); err != nil {
		return nil, errors.New("unable to unmarshal libmuttoncfg.json: " + err.Error())
	}
	return &cfg, nil
}
