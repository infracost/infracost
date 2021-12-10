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
	ctx         context.Context
	Logger      zerolog.Logger
	Config      *Config
	State       *State
	StartTime   int64
	contextVals map[string]interface{}
}

func NewRunContextFromEnv(rootCtx context.Context) (*RunContext, error) {
	cfg := DefaultConfig()
	err := cfg.LoadFromEnv()
	if err != nil {
		return EmptyRunContext(), err
	}

	state, _ := LoadState()

	c := &RunContext{
		ctx:         rootCtx,
		Logger:      configureLogger(),
		Config:      cfg,
		State:       state,
		contextVals: map[string]interface{}{},
		StartTime:   time.Now().Unix(),
	}

	c.loadInitialContextValues()

	return c, nil
}

func EmptyRunContext() *RunContext {
	return &RunContext{
		ctx: 			 context.Background(),
		Logger: configureLogger(),
		Config:      &Config{},
		State:       &State{},
		contextVals: map[string]interface{}{},
		StartTime:   time.Now().Unix(),
	}
}

func (c *RunContext) SetContextValue(key string, value interface{}) {
	c.contextVals[key] = value
}

func (c *RunContext) ContextValues() map[string]interface{} {
	return c.contextVals
}

func (c *RunContext) EventEnv() map[string]interface{} {
	env := c.contextVals
	env["installId"] = c.State.InstallID
	return env
}

func (c *RunContext) EventEnvWithProjectContexts(projectContexts []*ProjectContext) map[string]interface{} {
	env := c.contextVals
	env["installId"] = c.State.InstallID

	for _, projectContext := range projectContexts {
		if projectContext == nil {
			continue
		}

		for k, v := range projectContext.ContextValues() {
			if _, ok := env[k]; !ok {
				env[k] = make([]interface{}, 0)
			}
			env[k] = append(env[k].([]interface{}), v)
		}
	}

	return env
}

func (c *RunContext) loadInitialContextValues() {
	c.SetContextValue("version", baseVersion(version.Version))
	c.SetContextValue("fullVersion", version.Version)
	c.SetContextValue("isTest", IsTest())
	c.SetContextValue("isDev", IsDev())
	c.SetContextValue("os", runtime.GOOS)
	c.SetContextValue("ciPlatform", CIPlatform())
	c.SetContextValue("ciScript", ciScript())
	c.SetContextValue("ciPostCondition", os.Getenv("INFRACOST_CI_POST_CONDITION"))
	c.SetContextValue("ciPercentageThreshold", os.Getenv("INFRACOST_CI_PERCENTAGE_THRESHOLD"))
}

func configureLogger() zerolog.Logger {
	output := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	return zerolog.New(output).With().Timestamp().Logger()
}

func baseVersion(v string) string {
	return strings.SplitN(v, "+", 2)[0]
}
