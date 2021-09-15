package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func (c *Config) migrateConfiguration() error {
	// there are no migrations yet
	return nil
}

func (c *Config) migrateCredentials() error {
	oldPath := path.Join(userConfigDir(), "config.yml")
	credPath := CredentialsFilePath()

	if FileExists(oldPath) && !FileExists(credPath) {
		return c.migrateV0_7_17(oldPath, credPath)
	}

	if FileExists(credPath) {
		var content struct {
			Version string `yaml:"version"`
		}

		data, err := ioutil.ReadFile(credPath)
		if err != nil {
			return err
		}

		err = yaml.Unmarshal(data, &content)
		if err != nil {
			return err
		}

		if content.Version == "" {
			return c.migrateV0_9_4(credPath)
		}
	}

	return nil
}

func (c *Config) migrateV0_7_17(oldPath string, newPath string) error {
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
		c.Credentials.APIKey = oldCreds.APIKey
		c.Credentials.Version = "0.1"

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

func (c *Config) migrateV0_9_4(credPath string) error {
	log.Debugf("Migrating old credentials format to v0.1")

	// Use MapSlice to keep the order of the items, so we can always use the first one
	var oldCreds yaml.MapSlice

	data, err := ioutil.ReadFile(credPath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, &oldCreds)
	if err != nil {
		return err
	}

	err = os.Rename(credPath, fmt.Sprintf("%s.backup-%d", credPath, time.Now().Unix()))
	if err != nil {
		return err
	}

	// Get the first values
	var pricingAPIEndpoint string
	var apiKey string

	if len(oldCreds) > 0 {
		pricingAPIEndpoint = oldCreds[0].Key.(string)

		values, ok := oldCreds[0].Value.(yaml.MapSlice)
		if !ok {
			return errors.New("Invalid credentials format")
		}

		for _, item := range values {
			if item.Key.(string) == "api_key" {
				apiKey = item.Value.(string)
				break
			}
		}
	}

	c.Credentials.PricingAPIEndpoint = pricingAPIEndpoint
	c.Credentials.APIKey = apiKey
	c.Credentials.Version = "0.1"

	err = c.Credentials.Save()
	if err != nil {
		return err
	}

	log.Debug("Credentials successfully migrated")

	return nil
}
