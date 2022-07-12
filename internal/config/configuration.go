package config

import (
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
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
}

func loadConfiguration(cfg *Config) error {
	var err error

	err = cfg.migrateConfiguration()
	if err != nil {
		logrus.Debug("Error migrating configuration")
		logrus.Debug(err)
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

	if cfg.Configuration.DisableHCLParsing != nil {
		cfg.DisableHCLParsing = *cfg.Configuration.DisableHCLParsing
	}

	if cfg.Configuration.TLSInsecureSkipVerify != nil {
		cfg.TLSInsecureSkipVerify = cfg.Configuration.TLSInsecureSkipVerify
	}

	if cfg.TLSCACertFile == "" {
		cfg.TLSCACertFile = cfg.Configuration.TLSCACertFile
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
