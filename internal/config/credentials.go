package config

import (
	"io/ioutil"
	"os"
	"path"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type CredentialsProfileSpec struct {
	APIKey string `yaml:"apiKey"`
}

type CredentialsSpec map[string]CredentialsProfileSpec

var Credentials CredentialsSpec

func loadCredentials() error {
	var err error

	Credentials, err = readCredentialsFileIfExists()
	return err
}

func SaveCredentials() error {
	return writeCredentialsFile(Credentials)
}

func readCredentialsFileIfExists() (CredentialsSpec, error) {
	if !fileExists(CredentialsFilePath()) {
		return CredentialsSpec{}, nil
	}

	data, err := ioutil.ReadFile(CredentialsFilePath())
	if err != nil {
		return CredentialsSpec{}, err
	}

	var c CredentialsSpec

	err = yaml.Unmarshal(data, &c)

	return c, err
}

func writeCredentialsFile(c CredentialsSpec) error {
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

func migrateCredentials() error {
	oldPath := path.Join(userConfigDir(), "config.yml")

	if !fileExists(oldPath) {
		return nil
	}

	log.Debugf("Migrating old credentials from %s to %s", oldPath, CredentialsFilePath())

	data, err := ioutil.ReadFile(oldPath)
	if err != nil {
		return err
	}

	var oldCreds struct {
		APIKey string `yaml:"api_key"`
	}

	err = yaml.Unmarshal(data, &oldCreds)
	if err != nil {
		return err
	}

	if oldCreds.APIKey != "" {
		Credentials[Config.PricingAPIEndpoint] = CredentialsProfileSpec{
			APIKey: oldCreds.APIKey,
		}

		err = SaveCredentials()
		if err != nil {
			return err
		}

		err = os.Remove(oldPath)
		if err != nil {
			return err
		}

		log.Debug("Credentials successfully migrated")
	}

	return nil
}
