package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type TerraformProject struct {
	Name                string `yaml:"name,omitempty" ignored:"true"`
	Binary              string `yaml:"binary,omitempty" envconfig:"INFRACOST_TERRAFORM_BINARY"`
	Workspace           string `yaml:"workspace,omitempty" envconfig:"INFRACOST_TERRAFORM_WORKSPACE"`
	TerraformCloudHost  string `yaml:"terraform_cloud_host,omitempty" envconfig:"INFRACOST_TERRAFORM_CLOUD_HOST"`
	TerraformCloudToken string `yaml:"terraform_cloud_token,omitempty" envconfig:"INFRACOST_TERRAFORM_CLOUD_TOKEN"`
	UsageFile           string `yaml:"usage_file,omitempty" ignored:"true"`
	Dir                 string `yaml:"dir,omitempty" ignored:"true"`
	PlanFile            string `yaml:"plan_file,omitempty" ignored:"true"`
	JSONFile            string `yaml:"json_file,omitempty" ignored:"true"`
	PlanFlags           string `yaml:"plan_flags,omitempty" ignored:"true"`
	UseState            bool   `yaml:"use_state,omitempty" ignored:"true"`
}

func (p *TerraformProject) DisplayName() string {
	if p.Name != "" {
		return p.Name
	}

	if p.JSONFile != "" {
		return p.JSONFile
	}

	if p.PlanFile != "" {
		return p.PlanFile
	}

	if p.Dir != "" {
		return p.Dir
	}

	return "current directory"
}

type Projects struct {
	Terraform []*TerraformProject `yaml:"terraform,omitempty"`
}

type Output struct {
	Format      string   `yaml:"format,omitempty" ignored:"true"`
	Columns     []string `yaml:"columns,omitempty" ignored:"true"`
	ShowSkipped bool     `yaml:"show_skipped,omitempty" ignored:"true"`
	NoColor     bool     `yaml:"no_color,omitempty" ignored:"true"`
	Path        string   `yaml:"path,omitempty" ignored:"true"`
}

type Config struct { // nolint:golint
	Environment *Environment
	State       *State
	Credentials Credentials

	Version         string `yaml:"version,omitempty" ignored:"true"`
	LogLevel        string `yaml:"log_level,omitempty" envconfig:"INFRACOST_LOG_LEVEL"`
	NoColor         bool   `yaml:"no_color,omitempty" envconfig:"INFRACOST_NO_COLOR"`
	SkipUpdateCheck bool   `yaml:"skip_update_check,omitempty" envconfig:"INFRACOST_SKIP_UPDATE_CHECK"`

	APIKey                    string `envconfig:"INFRACOST_API_KEY"`
	PricingAPIEndpoint        string `yaml:"pricing_api_endpoint,omitempty" envconfig:"INFRACOST_PRICING_API_ENDPOINT"`
	DefaultPricingAPIEndpoint string `yaml:"default_pricing_api_endpoint,omitempty" envconfig:"INFRACOST_DEFAULT_PRICING_API_ENDPOINT"`
	DashboardAPIEndpoint      string `yaml:"dashboard_api_endpoint,omitempty" envconfig:"INFRACOST_DASHBOARD_API_ENDPOINT"`

	Projects Projects  `yaml:"projects" ignored:"true"`
	Outputs  []*Output `yaml:"outputs" ignored:"true"`
}

func init() {
	err := loadDotEnv()
	if err != nil {
		log.Fatal(err)
	}
}

func DefaultConfig() *Config {
	return &Config{
		Environment: NewEnvironment(),

		LogLevel: "",
		NoColor:  false,

		DefaultPricingAPIEndpoint: "https://pricing.api.infracost.io",
		PricingAPIEndpoint:        "https://pricing.api.infracost.io",
		DashboardAPIEndpoint:      "https://dashboard.api.infracost.io",

		Projects: Projects{
			Terraform: []*TerraformProject{
				{},
			},
		},
		Outputs: []*Output{
			{
				Format:  "table",
				Columns: []string{"NAME", "MONTHLY_QUANTITY", "UNIT", "PRICE", "HOURLY_COST", "MONTHLY_COST"},
			},
		},
	}
}

func (c *Config) LoadFromFile(configFile string) error {
	err := c.loadConfigFile(configFile)
	if err != nil {
		return err
	}

	err = c.LoadFromEnv()
	if err != nil {
		return err
	}

	if len(c.Projects.Terraform) > 0 {
		c.Environment.SetTerraformEnvironment(c.Projects.Terraform[0])
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

	err = loadState(c)
	if err != nil {
		logrus.Fatal(err)
	}
	c.Environment.InstallID = c.State.InstallID

	err = loadCredentials(c)
	if err != nil {
		logrus.Fatal(err)
	}

	return nil
}

func (c *Config) loadConfigFile(configFile string) error {
	if !fileExists(configFile) {
		return fmt.Errorf("Config file does not exist at %s", configFile)
	}

	c.Environment.HasConfigFile = true

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, c)
	if err != nil {
		return errors.New("Error parsing config YAML: " + strings.TrimPrefix(err.Error(), "yaml: "))
	}

	return nil
}

func (c *Config) loadEnvVars() error {
	err := envconfig.Process("", c)
	if err != nil {
		return err
	}

	for _, project := range c.Projects.Terraform {
		err = envconfig.Process("", project)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) ConfigureLogger() error {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		DisableColors: true,
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

	logrus.SetOutput(os.Stderr)

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

func loadDotEnv() error {
	envLocalPath := filepath.Join(RootDir(), ".env.local")
	if fileExists(envLocalPath) {
		err := godotenv.Load(envLocalPath)
		if err != nil {
			return err
		}
	}

	if fileExists(".env") {
		err := godotenv.Load()
		if err != nil {
			return err
		}
	}

	return nil
}
