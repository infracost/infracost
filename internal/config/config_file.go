package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"

	"github.com/mitchellh/go-homedir"
)

type StateFile struct {
	LatestReleaseVersion   string `json:"latestReleaseVersion"`
	LatestReleaseCheckedAt string `json:"latestReleaseCheckedAt"`
}

func ReadStateFile() (*StateFile, error) {
	var s *StateFile

	data, err := ioutil.ReadFile(stateFile())
	if err != nil {
		return s, err
	}

	err = json.Unmarshal(data, &s)
	return s, err
}

func WriteStateFile(s *StateFile) error {
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}

	err = os.MkdirAll(path.Dir(stateFile()), 0700)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(stateFile(), data, 0600)
}

func configDir() string {
	dir, _ := homedir.Expand("~/.config/infracost")
	return dir
}

func stateFile() string {
	return path.Join(configDir(), ".state.json")
}
