package config

import (
	"context"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/infracost/infracost/internal/version"
	"github.com/rs/zerolog"
)

type RunContext struct {
	ctx      context.Context
	logger   *zerolog.Logger
	config   *Config
	state    *State
	metadata map[string]interface{}
}

func NewRunContextFromEnv(ctx context.Context) (*RunContext, error) {
	cfg := DefaultConfig()
	err := cfg.LoadFromEnv()
	if err != nil {
		return EmptyRunContext(), err
	}

	state, _ := LoadState()

	c := &RunContext{
		ctx:      ctx,
		logger:   configureLogger(),
		config:   cfg,
		state:    state,
		metadata: map[string]interface{}{},
	}

	c.loadInitialContextValues()

	return c, nil
}

func EmptyRunContext() *RunContext {
	return &RunContext{
		ctx:      context.Background(),
		logger:   configureLogger(),
		config:   &Config{},
		state:    &State{},
		metadata: map[string]interface{}{},
	}
}

func (c *RunContext) WithLogger(logger *zerolog.Logger) *RunContext {
	return &RunContext{
		ctx:      c.ctx,
		logger:   logger,
		config:   c.config,
		state:    c.state,
		metadata: map[string]interface{}{},
	}
}

func (c *RunContext) Logger() *zerolog.Logger {
	return c.logger
}

func (c *RunContext) Config() *Config {
	return c.config
}

func (c *RunContext) State() *State {
	return c.state
}

func (c *RunContext) Metadata() map[string]interface{} {
	return c.metadata
}

func (c *RunContext) SetMetadata(key string, value interface{}) {
	c.metadata[key] = value
}

func (c *RunContext) loadInitialContextValues() {
	c.SetMetadata("installId", c.State().InstallID)
	c.SetMetadata("startTime", time.Now().Unix())
	c.SetMetadata("version", baseVersion(version.Version))
	c.SetMetadata("fullVersion", version.Version)
	c.SetMetadata("isTest", IsTest())
	c.SetMetadata("isDev", IsDev())
	c.SetMetadata("os", runtime.GOOS)
	c.SetMetadata("ciPlatform", CIPlatform())
	c.SetMetadata("ciScript", CIScript())
	c.SetMetadata("ciPostCondition", os.Getenv("INFRACOST_CI_POST_CONDITION"))
	c.SetMetadata("ciPercentageThreshold", os.Getenv("INFRACOST_CI_PERCENTAGE_THRESHOLD"))
}

func configureLogger() *zerolog.Logger {
	output := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	logger := zerolog.New(output).With().Timestamp().Logger()
	return &logger
}

func baseVersion(v string) string {
	return strings.SplitN(v, "+", 2)[0]
}
