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

type CredentialsProfileSpec struct {
	APIKey  string `yaml:"api_key"`
	Default bool   `yaml:"default,omitempty"`
}

type Credentials map[string]*CredentialsProfileSpec

func loadCredentials(cfg *Config) error {
	var err error

	cfg.Credentials, err = readCredentialsFileIfExists()
	if err != nil {
		return errors.New("Error parsing credentials YAML: " + strings.TrimPrefix(err.Error(), "yaml: "))
	}

	err = cfg.migrateCredentials()
	if err != nil {
		logrus.Debug("Error migrating credentials")
		logrus.Debug(err)
	}

	var profile *CredentialsProfileSpec
	var ok bool

	if cfg.PricingAPIEndpoint == "" {
		cfg.PricingAPIEndpoint = cfg.Credentials.GetDefaultPricingAPIEndpoint()
	}
	if cfg.PricingAPIEndpoint == "" {
		cfg.PricingAPIEndpoint = cfg.DefaultPricingAPIEndpoint
	}

	if cfg.PricingAPIEndpoint != "" {
		profile, ok = cfg.Credentials[cfg.PricingAPIEndpoint]
	}

	if ok && cfg.APIKey == "" {
		cfg.APIKey = profile.APIKey
	}

	return nil
}

func (c Credentials) Save() error {
	return writeCredentialsFile(c)
}

func (c Credentials) GetDefaultPricingAPIEndpoint() string {
	// Find the first key that's set as the default
	for k, v := range c {
		if v.Default {
			return k
		}
	}

	// If there's only one key in the map just return it
	if len(c) == 1 {
		for k := range c {
			return k
		}
	}

	return ""
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
