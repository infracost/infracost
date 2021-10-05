package config

import (
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var credentialsVersion = "0.1"

type Credentials struct {
	Version            string `yaml:"version"`
	APIKey             string `yaml:"api_key,omitempty"`
	PricingAPIEndpoint string `yaml:"pricing_api_endpoint,omitempty"`
}

func loadCredentials(cfg *Config) error {
	var err error

	err = cfg.migrateCredentials()
	if err != nil {
		logrus.Debug("Error migrating credentials")
		logrus.Debug(err)
	}

	cfg.Credentials, err = readCredentialsFileIfExists()
	if err != nil {
		return errors.New("Error parsing credentials YAML: " + strings.TrimPrefix(err.Error(), "yaml: "))
	}

	if cfg.PricingAPIEndpoint == "" {
		cfg.PricingAPIEndpoint = cfg.Credentials.PricingAPIEndpoint
	}
	if cfg.PricingAPIEndpoint == "" {
		cfg.PricingAPIEndpoint = cfg.DefaultPricingAPIEndpoint
	}

	if cfg.APIKey == "" {
		cfg.APIKey = cfg.Credentials.APIKey
	}

	return nil
}

func (c Credentials) Save() error {
	if c.Version == "" {
		c.Version = credentialsVersion
	}
	return writeCredentialsFile(c)
}

func readCredentialsFileIfExists() (Credentials, error) {
	if !FileExists(CredentialsFilePath()) {
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

func CredentialsFilePath() string {
	return path.Join(userConfigDir(), "credentials.yml")
}
