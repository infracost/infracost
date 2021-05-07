package config

import (
	"context"
	"os"
	"runtime"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/infracost/infracost/internal/version"
)

type RunContext struct {
	ctx               context.Context
	Config            *Config
	State             *State
	metadata          map[string]interface{}
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
		ctx:      rootCtx,
		Config:   cfg,
		State:    state,
		metadata: map[string]interface{}{},
	}

	c.loadInitialMetadata()

	return c, nil
}

func EmptyRunContext() *RunContext {
	return &RunContext{
		Config:   &Config{},
		State:    &State{},
		metadata: map[string]interface{}{},
	}
}

func (c *RunContext) SetMetadata(key string, value interface{}) {
	c.metadata[key] = value
}

func (c *RunContext) AllMetadata() map[string]interface{} {
	m := map[string]interface{}{
		"run": c.metadata,
	}
	if c.currentProjectCtx != nil {
		m["project"] = c.currentProjectCtx.metadata
	}

	return m
}

func (c *RunContext) SetCurrentProjectContext(ctx *ProjectContext) {
	c.currentProjectCtx = ctx
}

func (c *RunContext) loadInitialMetadata() {
	c.SetMetadata("runId", uuid.New().String())
	c.SetMetadata("version", baseVersion(version.Version))
	c.SetMetadata("fullVersion", version.Version)
	c.SetMetadata("isTest", IsTest())
	c.SetMetadata("isDev", IsDev())
	c.SetMetadata("os", runtime.GOOS)
	c.SetMetadata("ciPlatform", ciPlatform())
	c.SetMetadata("ciScript", ciScript())
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
