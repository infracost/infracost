package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func (c *Config) migrateCredentials() error {
	oldPath := path.Join(userConfigDir(), "config.yml")
	newPath := CredentialsFilePath()

	if !fileExists(oldPath) || fileExists(newPath) {
		return nil
	}

	log.Debugf("Migrating old credentials from %s to %s", oldPath, newPath)

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
		c.Credentials[c.PricingAPIEndpoint] = &CredentialsProfileSpec{
			APIKey: oldCreds.APIKey,
		}

		err = c.Credentials.Save()
		if err != nil {
			return err
		}

		err := os.Rename(oldPath, fmt.Sprintf("%s.backup-%d", oldPath, time.Now().Unix()))
		if err != nil {
			return err
		}

		log.Debug("Credentials successfully migrated")
	}

	return nil
}
