package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"

	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v2"
)

type StateFile struct {
	LatestReleaseVersion   string `json:"latestReleaseVersion"`
	LatestReleaseCheckedAt string `json:"latestReleaseCheckedAt"`
}

func ReadConfigFileIfExists() (ConfigSpec, error) {
	data, err := ioutil.ReadFile(ConfigFilePath())
	if os.IsNotExist(err) {
		return ConfigSpec{}, nil
	} else if err != nil {
		return ConfigSpec{}, err
	}

	var c ConfigSpec

	err = yaml.Unmarshal(data, &c)

	return c, err
}

func WriteConfigFile(c ConfigSpec) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	err = os.MkdirAll(path.Dir(ConfigFilePath()), 0700)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(ConfigFilePath(), data, 0600)
}

func mergeConfigFileIfExists(c *ConfigSpec) error {
	data, err := ioutil.ReadFile(ConfigFilePath())
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	return yaml.Unmarshal(data, c)
}

func ReadStateFileIfNotExists() (StateFile, error) {
	data, err := ioutil.ReadFile(StateFilePath())
	if os.IsNotExist(err) {
		return StateFile{}, nil
	} else if err != nil {
		return StateFile{}, err
	}

	var s StateFile
	err = json.Unmarshal(data, &s)

	return s, err
}

func WriteStateFile(s StateFile) error {
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}

	err = os.MkdirAll(path.Dir(StateFilePath()), 0700)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(StateFilePath(), data, 0600)
}

func configDir() string {
	dir, _ := homedir.Expand("~/.config/infracost")
	return dir
}

func ConfigFilePath() string { // nolint:golint
	return path.Join(configDir(), "config.yml")
}

func StateFilePath() string {
	return path.Join(configDir(), ".state.json")
}
