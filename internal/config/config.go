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

// ConfigSpec contains mapping of environment variable names to config values.
type ConfigSpec struct {
	NoColor                   bool
	LogLevel                  string `envconfig:"INFRACOST_LOG_LEVEL"  required:"false"`
	DefaultPricingAPIEndpoint string `envconfig:"DEFAULT_INFRACOST_PRICING_API_ENDPOINT" default:"https://pricing.api.infracost.io"`
	PricingAPIEndpoint        string `envconfig:"INFRACOST_PRICING_API_ENDPOINT" required:"true" default:"https://pricing.api.infracost.io"`
	DashboardAPIEndpoint      string `envconfig:"INFRACOST_DASHBOARD_API_ENDPOINT" required:"true" default:"https://dashboard.api.infracost.io"`
	APIKey                    string `envconfig:"INFRACOST_API_KEY"`
}

func (c *ConfigSpec) SetLogLevel(l string) error {
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

func (c *ConfigSpec) IsLogging() bool {
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

// loadConfig loads the config struct from environment variables.
func loadConfig() *ConfigSpec {
	var config ConfigSpec
	var err error

	config.NoColor = false

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

var Config = loadConfig()
