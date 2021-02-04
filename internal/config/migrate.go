package config

import (
	"io/ioutil"
	"os"
	"path"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func (c *Config) migrateCredentials() error {
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
		c.Credentials[c.PricingAPIEndpoint] = CredentialsProfileSpec{
			APIKey: oldCreds.APIKey,
		}

		err = c.Credentials.Save()
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
