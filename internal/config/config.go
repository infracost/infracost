package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type TerraformProjectSpec struct {
	Name                string `yaml:"name,omitempty" ignored:"true"`
	Binary              string `yaml:"binary,omitempty" envconfig:"TERRAFORM_BINARY"`
	Workspace           string `yaml:"workspace,omitempty" envconfig:"TERRAFORM_WORKSPACE"`
	TerraformCloudHost  string `yaml:"terraform_cloud_host,omitempty" envconfig:"TERRAFORM_CLOUD_HOST"`
	TerraformCloudToken string `yaml:"terraform_cloud_token,omitempty" envconfig:"TERRAFORM_CLOUD_TOKEN"`
	UsageFile           string `yaml:"usage_file,omitempty" ignored:"true"`
	Dir                 string `yaml:"dir,omitempty" ignored:"true"`
	PlanFile            string `yaml:"plan_file,omitempty" ignored:"true"`
	JSONFile            string `yaml:"json_file,omitempty" ignored:"true"`
	PlanFlags           string `yaml:"plan_flags,omitempty" ignored:"true"`
	UseState            bool   `yaml:"use_state,omitempty" ignored:"true"`
}

type ProjectSpec struct {
	Terraform []*TerraformProjectSpec `yaml:"terraform,omitempty"`
}

type OutputSpec struct {
	Format      string   `yaml:"format,omitempty" ignored:"true"`
	Columns     []string `yaml:"columns,omitempty" ignored:"true"`
	ShowSkipped bool     `yaml:"show_skipped,omitempty" ignored:"true"`
	Path        string   `yaml:"path,omitempty" ignored:"true"`
}

type ConfigSpec struct { // nolint:golint
	Version  string `yaml:"version,omitempty" ignored:"true"`
	LogLevel string `yaml:"log_level,omitempty" envconfig:"LOG_LEVEL"`
	NoColor  bool   `yaml:"no_color,omitempty" envconfig:"NO_COLOR"`

	APIKey                    string `envconfig:"API_KEY"`
	PricingAPIEndpoint        string `yaml:"pricing_api_endpoint,omitempty" envconfig:"PRICING_API_ENDPOINT"`
	DefaultPricingAPIEndpoint string `yaml:"default_pricing_api_endpoint,omitempty" envconfig:"DEFAULT_PRICING_API_ENDPOINT"`
	DashboardAPIEndpoint      string `yaml:"dashboard_api_endpoint,omitempty" envconfig:"DASHBOARD_API_ENDPOINT"`

	Projects ProjectSpec   `yaml:"projects" ignored:"true"`
	Outputs  []*OutputSpec `yaml:"outputs" ignored:"true"`
}

var Config *ConfigSpec

func LoadConfig(configFile string) {
	var err error

	err = loadConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}

	err = ConfigureLogger()
	if err != nil {
		log.Fatal(err)
	}

	err = loadCredentials()
	if err != nil {
		logrus.Fatal(err)
	}

	err = migrateCredentials()
	if err != nil {
		logrus.Debug("Error migrating credentials")
		logrus.Debug(err)
	}

	profile, ok := Credentials[Config.PricingAPIEndpoint]
	if ok && Config.APIKey == "" {
		Config.APIKey = profile.APIKey
	}

	if len(Config.Projects.Terraform) > 0 {
		LoadTerraformEnvironment(Config.Projects.Terraform[0])
	}
}

func loadConfig(configFile string) error {
	Config = defaultConfigSpec()

	err := mergeConfigFileIfExists(configFile)
	if err != nil {
		return err
	}

	err = loadDotEnv()
	if err != nil {
		return err
	}

	err = processEnvVars()
	if err != nil {
		return err
	}

	return nil
}

func defaultConfigSpec() *ConfigSpec {
	return &ConfigSpec{
		LogLevel: "",
		NoColor:  false,

		DefaultPricingAPIEndpoint: "https://pricing.api.infracost.io",
		PricingAPIEndpoint:        "https://pricing.api.infracost.io",
		DashboardAPIEndpoint:      "https://dashboard.api.infracost.io",

		Projects: ProjectSpec{
			Terraform: []*TerraformProjectSpec{
				{},
			},
		},
		Outputs: []*OutputSpec{
			{
				Format:  "table",
				Columns: []string{"NAME", "MONTHLY_QUANTITY", "UNIT", "PRICE", "HOURLY_COST", "MONTHLY_COST"},
			},
		},
	}
}

func mergeConfigFileIfExists(configFile string) error {
	if configFile != "" && !fileExists(configFile) {
		return fmt.Errorf("Config file does not exist at %s", configFile)
	}

	if configFile == "" {
		configFile = defaultConfigFilePath()

		if !fileExists(configFile) {
			return nil
		}
	}

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, Config)
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

func processEnvVars() error {
	err := envconfig.Process("INFRACOST_", Config)
	if err != nil {
		return err
	}

	for _, project := range Config.Projects.Terraform {
		err = envconfig.Process("INFRACOST_", project)
		if err != nil {
			return err
		}
	}

	return nil
}

func ConfigureLogger() error {
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

	if Config.LogLevel == "" {
		logrus.SetOutput(ioutil.Discard)
		return nil
	}

	logrus.SetOutput(os.Stderr)

	level, err := logrus.ParseLevel(Config.LogLevel)
	if err != nil {
		return err
	}

	logrus.SetLevel(level)

	return nil
}

func IsLogging() bool {
	return Config.LogLevel != ""
}

func defaultConfigFilePath() string { // nolint:golint
	cwd, err := os.Getwd()
	if err != nil {
		cwd = ""
	}

	return path.Join(cwd, "infracost.yml")
}
