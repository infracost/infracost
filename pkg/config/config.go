package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// ConfigSpec contains mapping of environment variable names to config values
type ConfigSpec struct {
	PriceListApiEndpoint string `envconfig:"PLAN_COSTS_PRICE_LIST_API_ENDPOINT"  required:"true"  default:"http://localhost:4000/graphql"`
}

// loadConfig loads the config struct from environment variables
func loadConfig() (*ConfigSpec, error) {
	var config ConfigSpec
	var err error

	err = godotenv.Load(".env.local")
	if (err != nil) {
		return nil, err
	}
	err = godotenv.Load()
	if (err != nil) {
		return nil, err
	}

	err = envconfig.Process("", &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

var Config, _ = loadConfig()
