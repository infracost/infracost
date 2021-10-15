package config

import (
	"encoding/json"
	"os"
	"path"

	"github.com/google/uuid"
)

type State struct {
	InstallID              string `json:"installId"`
	LatestReleaseVersion   string `json:"latestReleaseVersion"`
	LatestReleaseCheckedAt string `json:"latestReleaseCheckedAt"`
}

func LoadState() (*State, error) {
	state, err := readStateFileIfExists()
	if err != nil {
		return state, err
	}

	if state.InstallID == "" {
		state.InstallID = uuid.New().String()
		err = state.Save()
		if err != nil {
			return state, err
		}
	}

	return state, nil
}

func (s *State) Save() error {
	return writeStateFile(s)
}

func readStateFileIfExists() (*State, error) {
	if !FileExists(stateFilePath()) {
		return &State{}, nil
	}

	data, err := os.ReadFile(stateFilePath())
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

	return os.WriteFile(stateFilePath(), data, 0600)
}

func stateFilePath() string {
	return path.Join(userConfigDir(), ".state.json")
}
