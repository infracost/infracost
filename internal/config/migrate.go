package config

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/infracost/infracost/internal/logging"
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

		data, err := os.ReadFile(credPath)
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
	logging.Logger.Debug().Msgf("Migrating old credentials from %s to %s", oldPath, newPath)

	data, err := os.ReadFile(oldPath)
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

		logging.Logger.Debug().Msg("Credentials successfully migrated")
	}

	return nil
}

func (c *Config) migrateV0_9_4(credPath string) error {
	logging.Logger.Debug().Msgf("Migrating old credentials format to v0.1")

	// Use MapSlice to keep the order of the items, so we can always use the first one
	var oldCreds yaml.MapSlice

	data, err := os.ReadFile(credPath)
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

	logging.Logger.Debug().Msg("Credentials successfully migrated")

	return nil
}
