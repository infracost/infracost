package config

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type CredentialsProfileSpec struct {
	APIKey string `yaml:"apiKey"`
}

type Credentials map[string]CredentialsProfileSpec

func loadCredentials(cfg *Config) error {
	var err error

	cfg.Credentials, err = readCredentialsFileIfExists()
	if err != nil {
		return err
	}

	err = cfg.migrateCredentials()
	if err != nil {
		logrus.Debug("Error migrating credentials")
		logrus.Debug(err)
	}

	profile, ok := cfg.Credentials[cfg.PricingAPIEndpoint]
	if ok && cfg.APIKey == "" {
		cfg.APIKey = profile.APIKey
	}

	return nil
}

func (c Credentials) Save() error {
	return writeCredentialsFile(c)
}

func readCredentialsFileIfExists() (Credentials, error) {
	if !fileExists(CredentialsFilePath()) {
		return Credentials{}, nil
	}

	data, err := ioutil.ReadFile(CredentialsFilePath())
	if err != nil {
		return Credentials{}, err
	}

	var c Credentials

	err = yaml.Unmarshal(data, &c)

	return c, err
}

func writeCredentialsFile(c Credentials) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	err = os.MkdirAll(path.Dir(CredentialsFilePath()), 0700)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(CredentialsFilePath(), data, 0600)
}

func CredentialsFilePath() string { // nolint:golint
	return path.Join(userConfigDir(), "credentials.yml")
}
