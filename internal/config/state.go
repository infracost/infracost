package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"

	"github.com/google/uuid"
)

type State struct {
	InstallID              string `json:"installId"`
	LatestReleaseVersion   string `json:"latestReleaseVersion"`
	LatestReleaseCheckedAt string `json:"latestReleaseCheckedAt"`
}

func loadState(cfg *Config) error {
	var err error

	cfg.State, err = readStateFileIfExists()
	if err != nil {
		return err
	}

	if cfg.State.InstallID == "" {
		cfg.State.InstallID = uuid.New().String()
		err = cfg.State.Save()
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *State) Save() error {
	return writeStateFile(s)
}

func readStateFileIfExists() (*State, error) {
	if !fileExists(stateFilePath()) {
		return &State{}, nil
	}

	data, err := ioutil.ReadFile(stateFilePath())
	if err != nil {
		return &State{}, err
	}

	var s State
	err = json.Unmarshal(data, &s)

	return &s, err
}

func writeStateFile(s *State) error {
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
