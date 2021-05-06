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

type RunMetadata struct {
	Version                     string   `json:"version"`
	FullVersion                 string   `json:"fullVersion"`
	IsTest                      bool     `json:"isTest"`
	IsDev                       bool     `json:"isDev"`
	InstallID                   string   `json:"installId"`
	IsDefaultPricingAPIEndpoint bool     `json:"isDefaultPricingAPIEndpoint"`
	OS                          string   `json:"os"`
	CIPlatform                  string   `json:"ciPlatform,omitempty"`
	CIScript                    string   `json:"ciScript,omitempty"`
	Command                     string   `json:"command"`
	Flags                       []string `json:"flags"`
	OutputFormat                string   `json:"outputFormat"`
	HasConfigFile               bool     `json:"hasConfigFile"`
}

type RunContext struct {
	ctx      context.Context
	Config   *Config
	State    *State
	RunID    string
	Metadata RunMetadata
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

	metadata := loadInitialRunMetadata()

	return &RunContext{
		ctx:      rootCtx,
		Config:   cfg,
		State:    state,
		RunID:    uuid.New().String(),
		Metadata: metadata,
	}, nil
}

func EmptyRunContext() *RunContext {
	return &RunContext{
		Config:   &Config{},
		State:    &State{},
		Metadata: RunMetadata{},
	}
}

func loadInitialRunMetadata() RunMetadata {
	return RunMetadata{
		Version:     baseVersion(version.Version),
		FullVersion: version.Version,
		IsTest:      isTest(),
		IsDev:       isDev(),
		OS:          runtime.GOOS,
		CIPlatform:  ciPlatform(),
		CIScript:    ciScript(),
	}
}

func baseVersion(v string) string {
	return strings.SplitN(v, "+", 2)[0]
}

func terraformVersion(full string) string {
	p := strings.Split(full, " ")
	if len(p) > 1 {
		return p[len(p)-1]
	}

	return ""
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

func isTest() bool {
	return os.Getenv("INFRACOST_ENV") == "test" || strings.HasSuffix(os.Args[0], ".test")
}

func isDev() bool {
	return os.Getenv("INFRACOST_ENV") == "dev"
}
