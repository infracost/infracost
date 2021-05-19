package config

import (
	"context"
	"os"
	"runtime"
	"sort"
	"strings"

	"github.com/infracost/infracost/internal/version"
)

type RunContext struct {
	ctx               context.Context
	Config            *Config
	State             *State
	contextVals       map[string]interface{}
	currentProjectCtx *ProjectContext
}

func NewRunContextFromEnv(rootCtx context.Context) (*RunContext, error) {
	cfg := DefaultConfig()
	err := cfg.LoadFromEnv()
	if err != nil {
		return nil, err
	}

	state, err := LoadState()
	if err != nil {
		return nil, err
	}

	c := &RunContext{
		ctx:         rootCtx,
		Config:      cfg,
		State:       state,
		contextVals: map[string]interface{}{},
	}

	c.loadInitialContextValues()

	return c, nil
}

func EmptyRunContext() *RunContext {
	return &RunContext{
		Config:      &Config{},
		State:       &State{},
		contextVals: map[string]interface{}{},
	}
}

func (c *RunContext) SetContextValue(key string, value interface{}) {
	c.contextVals[key] = value
}

func (c *RunContext) ContextValues() map[string]interface{} {
	return c.contextVals
}

func (c *RunContext) ContextValuesWithCurrentProject() map[string]interface{} {
	m := c.contextVals
	if c.currentProjectCtx != nil {
		for k, v := range c.currentProjectCtx.contextVals {
			m[k] = v
		}
	}

	return m
}

func (c *RunContext) EventEnv() map[string]interface{} {
	return c.EventEnvWithProjectContexts([]*ProjectContext{c.currentProjectCtx})
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

func (c *RunContext) SetCurrentProjectContext(ctx *ProjectContext) {
	c.currentProjectCtx = ctx
}

func (c *RunContext) loadInitialContextValues() {
	c.SetContextValue("version", baseVersion(version.Version))
	c.SetContextValue("fullVersion", version.Version)
	c.SetContextValue("isTest", IsTest())
	c.SetContextValue("isDev", IsDev())
	c.SetContextValue("os", runtime.GOOS)
	c.SetContextValue("ciPlatform", ciPlatform())
	c.SetContextValue("ciScript", ciScript())
}

func baseVersion(v string) string {
	return strings.SplitN(v, "+", 2)[0]
}

func ciScript() string {
	if IsTruthy(os.Getenv("INFRACOST_CI_DIFF")) {
		return "ci-diff"
	} else if IsTruthy(os.Getenv("INFRACOST_CI_ATLANTIS_DIFF")) {
		return "ci-atlantis-diff"
	}

	return ""
}

func ciPlatform() string {
	if IsTruthy(os.Getenv("GITHUB_ACTIONS")) {
		return "github_actions"
	} else if IsTruthy(os.Getenv("GITLAB_CI")) {
		return "gitlab_ci"
	} else if IsTruthy(os.Getenv("CIRCLECI")) {
		return "circleci"
	} else {
		envKeys := os.Environ()
		sort.Strings(envKeys)
		for _, k := range envKeys {
			if strings.HasPrefix(k, "ATLANTIS_") {
				return "atlantis"
			} else if strings.HasPrefix(k, "BITBUCKET_") {
				return "bitbucket"
			} else if strings.HasPrefix(k, "JENKINS_") {
				return "jenkins"
			} else if strings.HasPrefix(k, "CONCOURSE_") {
				return "concourse"
			}
		}
		if IsTruthy(os.Getenv("CI")) {
			return "ci"
		}
	}

	return ""
}
