package config

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/infracost/infracost/internal/logging"
)

var configurationVersion = "0.1"

type Configuration struct {
	Version               string `yaml:"version"`
	Currency              string `yaml:"currency,omitempty"`
	EnableDashboard       *bool  `yaml:"enable_dashboard,omitempty"`
	DisableHCLParsing     *bool  `yaml:"disable_hcl_parsing,omitempty"`
	TLSInsecureSkipVerify *bool  `yaml:"tls_insecure_skip_verify,omitempty"`
	TLSCACertFile         string `yaml:"tls_ca_cert_file,omitempty"`
	EnableCloud           *bool  `yaml:"enable_cloud"`
	EnableCloudUpload     *bool  `yaml:"enable_cloud_upload"`

	ProductionFilters []ProductionFilter `yaml:"production_filters,omitempty"`
}

// ProductionFilter is a filter for production/non-production paths..
type ProductionFilter struct {
	ID      string `yaml:"id"`
	Type    string `yaml:"type"`
	Include bool   `yaml:"include"`
	Value   string `yaml:"value"`
}

func loadConfiguration(cfg *Config) error {
	var err error

	err = cfg.migrateConfiguration()
	if err != nil {
		logging.Logger.Debug().Err(err).Msg("error migrating configuration")
	}

	cfg.Configuration, err = readConfigurationFileIfExists()
	if err != nil {
		return errors.New("Error parsing configuration YAML: " + strings.TrimPrefix(err.Error(), "yaml: "))
	}

	if cfg.Currency == "" {
		cfg.Currency = cfg.Configuration.Currency
	}
	if cfg.Currency == "" {
		cfg.Currency = "USD"
	}

	if cfg.Configuration.EnableDashboard != nil {
		cfg.EnableDashboard = *cfg.Configuration.EnableDashboard
	}

	if cfg.Configuration.EnableCloud != nil {
		cfg.EnableCloud = cfg.Configuration.EnableCloud
	}

	if cfg.Configuration.EnableCloudUpload != nil {
		cfg.EnableCloudUpload = cfg.Configuration.EnableCloudUpload
	}

	if cfg.Configuration.DisableHCLParsing != nil {
		cfg.DisableHCLParsing = *cfg.Configuration.DisableHCLParsing
	}

	if cfg.Configuration.TLSInsecureSkipVerify != nil {
		cfg.TLSInsecureSkipVerify = cfg.Configuration.TLSInsecureSkipVerify
	}

	if cfg.TLSCACertFile == "" {
		cfg.TLSCACertFile = cfg.Configuration.TLSCACertFile
	}

	if cfg.Configuration.ProductionFilters != nil {
		for i := range cfg.Projects {
			if cfg.Projects[i].Metadata == nil {
				cfg.Projects[i].Metadata = map[string]string{}
			}

			// Only set the isProduction metadata if it doesn't already exist
			// This allows it to be overridden by a project level config
			if _, ok := cfg.Projects[i].Metadata["isProduction"]; !ok && !IsTest() {
				cfg.Projects[i].Metadata["isProduction"] = fmt.Sprintf("%t", cfg.IsProduction(cfg.Projects[i].Name))
			}
		}
	}

	return nil
}

func (c Configuration) Save() error {
	if c.Version == "" {
		c.Version = configurationVersion
	}
	return writeConfigurationFile(c)
}

func readConfigurationFileIfExists() (Configuration, error) {
	if !FileExists(ConfigurationFilePath()) {
		return Configuration{}, nil
	}

	data, err := os.ReadFile(ConfigurationFilePath())
	if err != nil {
		return Configuration{}, err
	}

	var c Configuration

	err = yaml.Unmarshal(data, &c)

	return c, err
}

func writeConfigurationFile(c Configuration) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	err = os.MkdirAll(path.Dir(ConfigurationFilePath()), 0700)
	if err != nil {
		return err
	}

	return os.WriteFile(ConfigurationFilePath(), data, 0600)
}

func ConfigurationFilePath() string {
	return path.Join(userConfigDir(), "configuration.yml")
}

// IsProduction returns true if the project is production.
func (c *Config) IsProduction(value string) bool {
	matchesProduction := false

	for _, filter := range c.Configuration.ProductionFilters {
		if filter.Type != "PROJECT" {
			continue
		}

		isMatch := matchesWildcard(filter.Value, value)

		if filter.Include && isMatch {
			matchesProduction = true
		}

		// If it matches a non-production filter, it's definitely not production
		if !filter.Include && isMatch {
			return false
		}
	}

	// Project is production only if it matched a production filter
	// and didn't match any non-production filters
	return matchesProduction
}

// matchesWildcard checks if the pattern matches the string, supporting * as a wildcard
// that matches any number of characters
func matchesWildcard(pattern, s string) bool {
	// If there's no wildcard, just do a sub string comparison
	if !strings.Contains(pattern, "*") {
		return strings.Contains(s, pattern)
	}

	parts := strings.Split(pattern, "*")

	if parts[0] != "" {
		if !strings.HasPrefix(s, parts[0]) {
			return false
		}
		s = s[len(parts[0]):]
	}

	if parts[len(parts)-1] != "" {
		if !strings.HasSuffix(s, parts[len(parts)-1]) {
			return false
		}
		s = s[:len(s)-len(parts[len(parts)-1])]
		parts = parts[:len(parts)-1]
	}

	for _, part := range parts[1:] {
		if part == "" {
			continue
		}

		idx := strings.Index(s, part)
		if idx == -1 {
			return false
		}

		s = s[idx+len(part):]
	}

	return true
}
