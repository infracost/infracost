package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
)

type StateSpec struct {
	LatestReleaseVersion   string `json:"latestReleaseVersion"`
	LatestReleaseCheckedAt string `json:"latestReleaseCheckedAt"`
}

var State *StateSpec

func init() {
	err := loadState()
	if err != nil {
		log.Fatal(err)
	}
}

func loadState() error {
	var err error

	State, err = readStateFileIfExists()
	return err
}

func SaveState() error {
	return writeStateFile(State)
}

func readStateFileIfExists() (*StateSpec, error) {
	if !fileExists(stateFilePath()) {
		return &StateSpec{}, nil
	}

	data, err := ioutil.ReadFile(stateFilePath())
	if err != nil {
		return &StateSpec{}, err
	}

	var s StateSpec
	err = json.Unmarshal(data, &s)

	return &s, err
}

func writeStateFile(s *StateSpec) error {
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}

	err = os.MkdirAll(path.Dir(stateFilePath()), 0700)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(stateFilePath(), data, 0600)
}

func stateFilePath() string {
	return path.Join(userConfigDir(), ".state.json")
}
