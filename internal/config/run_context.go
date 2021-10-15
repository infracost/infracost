package config

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/infracost/infracost/internal/version"
)

type RunContext struct {
	ctx               context.Context
	Config            *Config
	State             *State
	contextVals       map[string]interface{}
	currentProjectCtx *ProjectContext
	StartTime         int64
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

// setProjectContextValue Set context value into currentProjectContext
func (c *RunContext) setProjectContextValue(key string, value interface{}) {
	c.currentProjectCtx.SetContextValue(key, value)
}

func (c *RunContext) loadInitialContextValues() {
	c.SetContextValue("version", baseVersion(version.Version))
	c.SetContextValue("fullVersion", version.Version)
	c.SetContextValue("isTest", IsTest())
	c.SetContextValue("isDev", IsDev())
	c.SetContextValue("os", runtime.GOOS)
	c.SetContextValue("ciPlatform", ciPlatform())
	c.SetContextValue("ciScript", ciScript())
	c.SetContextValue("ciPostCondition", os.Getenv("INFRACOST_CI_POST_CONDITION"))
	c.SetContextValue("ciPercentageThreshold", os.Getenv("INFRACOST_CI_PERCENTAGE_THRESHOLD"))
}

type ProjectContexter interface {
	ProjectContext() map[string]interface{}
}

func (c *RunContext) SetProjectContextFrom(d ProjectContexter) {
	m := d.ProjectContext()
	for k, v := range m {
		c.setProjectContextValue(k, v)
	}
}

func baseVersion(v string) string {
	return strings.SplitN(v, "+", 2)[0]
}

func ciScript() string {
	if IsEnvPresent("INFRACOST_CI_DIFF") {
		return "ci-diff"
	} else if IsEnvPresent("INFRACOST_CI_ATLANTIS_DIFF") {
		return "ci-atlantis-diff"
	} else if IsEnvPresent("INFRACOST_CI_JENKINS_DIFF") {
		return "ci-jenkins-diff"
	}

	return ""
}

func ciPlatform() string {
	if IsEnvPresent("GITHUB_ACTIONS") {
		return "github_actions"
	} else if IsEnvPresent("GITLAB_CI") {
		return "gitlab_ci"
	} else if IsEnvPresent("CIRCLECI") {
		return "circleci"
	} else if IsEnvPresent("JENKINS_HOME") {
		return "jenkins"
	} else if IsEnvPresent("BUILDKITE") {
		return "buildkite"
	} else if IsEnvPresent("SYSTEM_COLLECTIONURI") {
		return fmt.Sprintf("azure_devops_%s", os.Getenv("BUILD_REPOSITORY_PROVIDER"))
	} else if IsEnvPresent("TFC_RUN_ID") {
		return "tfc"
	} else if IsEnvPresent("ENV0_ENVIRONMENT_ID") {
		return "env0"
	} else if IsEnvPresent("SCALR_RUN_ID") {
		return "scalr"
	} else {
		envKeys := os.Environ()
		sort.Strings(envKeys)
		for _, k := range envKeys {
			if strings.HasPrefix(k, "ATLANTIS_") {
				return "atlantis"
			} else if strings.HasPrefix(k, "BITBUCKET_") {
				return "bitbucket"
			} else if strings.HasPrefix(k, "CONCOURSE_") {
				return "concourse"
			} else if strings.HasPrefix(k, "SPACELIFT_") {
				return "spacelift"
			} else if strings.HasPrefix(k, "HARNESS_") {
				return "harness"
			}
		}
		if IsEnvPresent("CI") {
			return "ci"
		}
	}

	return ""
}
