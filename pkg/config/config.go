package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

// ConfigSpec contains mapping of environment variable names to config values
type ConfigSpec struct {
	Logger *logrus.Logger
	ApiUrl string `envconfig:"PLANCOSTS_API_URL"  required:"true"  default:"https://aws-prices-graphql.alistair.scot"`
}

func (c *ConfigSpec) SetLogger(logger *logrus.Logger) {
	c.Logger = logger
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

	if fileExists(".env.local") {
		err = godotenv.Load(".env.local")
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
	return &config
}

var Config = loadConfig()
