package config

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

type Project struct {
	Path                string `yaml:"path,omitempty" ignored:"true"`
	TerraformPlanFlags  string `yaml:"terraform_plan_flags,omitempty" ignored:"true"`
	TerraformBinary     string `yaml:"terraform_binary,omitempty" envconfig:"INFRACOST_TERRAFORM_BINARY"`
	TerraformWorkspace  string `yaml:"terraform_workspace,omitempty" envconfig:"INFRACOST_TERRAFORM_WORKSPACE"`
	TerraformCloudHost  string `yaml:"terraform_cloud_host,omitempty" envconfig:"INFRACOST_TERRAFORM_CLOUD_HOST"`
	TerraformCloudToken string `yaml:"terraform_cloud_token,omitempty" envconfig:"INFRACOST_TERRAFORM_CLOUD_TOKEN"`
	UsageFile           string `yaml:"usage_file,omitempty" ignored:"true"`
	TerraformUseState   bool   `yaml:"terraform_use_state,omitempty" ignored:"true"`
}

type Config struct {
	Credentials   Credentials
	Configuration Configuration

	Version         string `yaml:"version,omitempty" ignored:"true"`
	LogLevel        string `yaml:"log_level,omitempty" envconfig:"INFRACOST_LOG_LEVEL"`
	NoColor         bool   `yaml:"no_color,omitempty" envconfig:"INFRACOST_NO_COLOR"`
	SkipUpdateCheck bool   `yaml:"skip_update_check,omitempty" envconfig:"INFRACOST_SKIP_UPDATE_CHECK"`

	APIKey                    string `envconfig:"INFRACOST_API_KEY"`
	PricingAPIEndpoint        string `yaml:"pricing_api_endpoint,omitempty" envconfig:"INFRACOST_PRICING_API_ENDPOINT"`
	DefaultPricingAPIEndpoint string `yaml:"default_pricing_api_endpoint,omitempty" envconfig:"INFRACOST_DEFAULT_PRICING_API_ENDPOINT"`
	DashboardAPIEndpoint      string `yaml:"dashboard_api_endpoint,omitempty" envconfig:"INFRACOST_DASHBOARD_API_ENDPOINT"`
	EnableDashboard           bool   `yaml:"enable_dashboard,omitempty" envconfig:"INFRACOST_ENABLE_DASHBOARD"`

	Currency string `envconfig:"INFRACOST_CURRENCY"`

	Projects      []*Project `yaml:"projects" ignored:"true"`
	Format        string     `yaml:"format,omitempty" ignored:"true"`
	ShowSkipped   bool       `yaml:"show_skipped,omitempty" ignored:"true"`
	SyncUsageFile bool       `yaml:"sync_usage_file,omitempty" ignored:"true"`
	Fields        []string   `yaml:"fields,omitempty" ignored:"true"`

	// for testing
	EventsDisabled       bool
	LogWriter            io.Writer
	LogDisableTimestamps bool
}

func init() {
	err := loadDotEnv()
	if err != nil {
		log.Fatal(err)
	}
}

func DefaultConfig() *Config {
	return &Config{
		LogLevel: "",
		NoColor:  false,

		DefaultPricingAPIEndpoint: "https://pricing.api.infracost.io",
		PricingAPIEndpoint:        "",
		DashboardAPIEndpoint:      "https://dashboard.api.infracost.io",

		Projects: []*Project{{}},

		Format: "table",
		Fields: []string{"monthlyQuantity", "unit", "monthlyCost"},

		EventsDisabled: IsTest(),
	}
}

func (c *Config) LoadFromConfigFile(path string) error {
	cfgFile, err := LoadConfigFile(path)
	if err != nil {
		return err
	}

	c.Projects = cfgFile.Projects

	// Reload the environment to overwrite any of the config file configs
	err = c.LoadFromEnv()
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) LoadFromEnv() error {
	err := c.loadEnvVars()
	if err != nil {
		return err
	}

	err = c.ConfigureLogger()
	if err != nil {
		return err
	}

	err = loadCredentials(c)
	if err != nil {
		return err
	}

	err = loadConfiguration(c)
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) loadEnvVars() error {
	err := envconfig.Process("", c)
	if err != nil {
		return err
	}

	for _, project := range c.Projects {
		err = envconfig.Process("", project)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) ConfigureLogger() error {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:    true,
		DisableTimestamp: c.LogDisableTimestamps,
		DisableColors:    true,
		SortingFunc: func(keys []string) {
			// Put message at the end
			for i, key := range keys {
				if key == "msg" && i != len(keys)-1 {
					keys[i], keys[len(keys)-1] = keys[len(keys)-1], keys[i]
					break
				}
			}
		},
	})

	if c.LogLevel == "" {
		logrus.SetOutput(ioutil.Discard)
		return nil
	}

	if c.LogWriter != nil {
		logrus.SetOutput(c.LogWriter)
	} else {
		logrus.SetOutput(os.Stderr)
	}

	level, err := logrus.ParseLevel(c.LogLevel)
	if err != nil {
		return err
	}

	logrus.SetLevel(level)

	return nil
}

func (c *Config) IsLogging() bool {
	return c.LogLevel != ""
}

func IsTest() bool {
	return os.Getenv("INFRACOST_ENV") == "test" || strings.HasSuffix(os.Args[0], ".test")
}

func IsDev() bool {
	return os.Getenv("INFRACOST_ENV") == "dev"
}

func loadDotEnv() error {
	envLocalPath := filepath.Join(RootDir(), ".env.local")
	if FileExists(envLocalPath) {
		err := godotenv.Load(envLocalPath)
		if err != nil {
			return err
		}
	}

	if FileExists(".env") {
		err := godotenv.Load()
		if err != nil {
			return err
		}
	}

	return nil
}
