package config

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type ConfigFileSpec struct { // nolint:golint
	Projects []*Project `yaml:"projects" ignored:"true"`
}

func LoadConfigFile(path string) (ConfigFileSpec, error) {
	cfgFile := ConfigFileSpec{}

	if !fileExists(path) {
		return cfgFile, fmt.Errorf("Config file does not exist at %s", path)
	}

	rawCfgFile, err := ioutil.ReadFile(path)
	if err != nil {
		return cfgFile, err
	}

	err = yaml.Unmarshal(rawCfgFile, &cfgFile)
	if err != nil {
		return cfgFile, errors.New("Error parsing config YAML: " + strings.TrimPrefix(err.Error(), "yaml: "))
	}

	return cfgFile, nil
}
