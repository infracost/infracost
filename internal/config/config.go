package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/infracost/infracost/internal/version"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

// Spec contains mapping of environment variable names to config values
type ConfigSpec struct { // nolint:golint
	NoColor                   bool   `yaml:"no_color,omitempty"`
	LogLevel                  string `yaml:"log_level,omitempty" envconfig:"INFRACOST_LOG_LEVEL"`
	DefaultPricingAPIEndpoint string `yaml:"default_pricing_api_endpoint,omitempty" envconfig:"INFRACOST_DEFAULT_PRICING_API_ENDPOINT"`
	PricingAPIEndpoint        string `yaml:"pricing_api_endpoint,omitempty" envconfig:"INFRACOST_PRICING_API_ENDPOINT"`
	DashboardAPIEndpoint      string `yaml:"dashboard_api_endpoint,omitempty" envconfig:"INFRACOST_DASHBOARD_API_ENDPOINT"`
	APIKey                    string `yaml:"api_key,omitempty" envconfig:"INFRACOST_API_KEY"`
}

var Config *ConfigSpec

func init() {
	log.SetFlags(0)

	Config = loadConfig()
}

func defaultConfigSpec() ConfigSpec {
	return ConfigSpec{
		NoColor:                   false,
		DefaultPricingAPIEndpoint: "https://pricing.api.infracost.io",
		PricingAPIEndpoint:        "https://pricing.api.infracost.io",
		DashboardAPIEndpoint:      "https://dashboard.api.infracost.io",
	}
}

func (c ConfigSpec) SetLogLevel(l string) error {
	c.LogLevel = l

	// Disable logging if no log level is set
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

func (c ConfigSpec) IsLogging() bool {
	return c.LogLevel != ""
}

func LogSortingFunc(keys []string) {
	// Put message at the end
	for i, key := range keys {
		if key == "msg" && i != len(keys)-1 {
			keys[i], keys[len(keys)-1] = keys[len(keys)-1], keys[i]
			break
		}
	}
}

func RootDir() string {
	_, b, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(b), "../..")
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}

// loadConfig loads the config spec from the config file and environment variables.
// Config is loaded in the following order, with any later ones overriding the previous ones:
//   * Default values
//   * Config file
//   * .env
//   * .env.local
//   * ENV variables
//   * Any command line flags (e.g. --log-level)
func loadConfig() *ConfigSpec {
	config := defaultConfigSpec()

	err := mergeConfigFileIfExists(&config)
	if err != nil {
		log.Fatal(err)
	}

	envLocalPath := filepath.Join(RootDir(), ".env.local")
	if fileExists(envLocalPath) {
		err = godotenv.Load(envLocalPath)
		if err != nil {
			log.Fatal(err)
		}
	}

	if fileExists(".env") {
		err = godotenv.Load()
		if err != nil {
			log.Fatal(err)
		}
	}

	err = envconfig.Process("", &config)
	if err != nil {
		log.Fatal(err)
	}

	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		DisableColors: true,
		SortingFunc:   LogSortingFunc,
	})

	err = config.SetLogLevel(config.LogLevel)
	if err != nil {
		log.Fatal(err)
	}

	return &config
}

func GetUserAgent() string {
	userAgent := "infracost"

	if version.Version != "" {
		userAgent += fmt.Sprintf("-%s", version.Version)
	}

	infracostEnv := getInfracostEnv()
	if infracostEnv != "" {
		userAgent += fmt.Sprintf(" (%s)", infracostEnv)
	}

	return userAgent
}

func getInfracostEnv() string {
	if IsTest() {
		return "test"
	} else if IsDev() {
		return "dev"
	} else if IsTruthy(os.Getenv("GITHUB_ACTIONS")) {
		return "github_actions"
	} else if IsTruthy(os.Getenv("GITLAB_CI")) {
		return "gitlab_ci"
	} else if IsTruthy(os.Getenv("CIRCLECI")) {
		return "circleci"
	}

	return ""
}

func IsTest() bool {
	return os.Getenv("INFRACOST_ENV") == "test" || strings.HasSuffix(os.Args[0], ".test")
}

func IsDev() bool {
	return os.Getenv("INFRACOST_ENV") == "dev"
}

func IsTruthy(s string) bool {
	return s == "1" || strings.EqualFold(s, "true")
}
