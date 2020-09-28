package config

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

// ConfigSpec contains mapping of environment variable names to config values
type ConfigSpec struct {
	NoColor  bool
	ApiUrl   string `envconfig:"INFRACOST_API_URL"  required:"true"  default:"https://pricing.infracost.io"`
	LogLevel string `envconfig:"INFRACOST_LOG_LEVEL"  required:"false"`
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

// loadConfig loads the config struct from environment variables
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

var Config = loadConfig()
