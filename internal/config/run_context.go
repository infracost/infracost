package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/infracost/infracost/internal/version"
)

type RunContext struct {
	ctx         context.Context
	uuid        uuid.UUID
	Config      *Config
	State       *State
	contextVals map[string]interface{}
	StartTime   int64

	OutWriter io.Writer
	ErrWriter io.Writer
	Exit      func(code int)
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
		OutWriter:   os.Stdout,
		ErrWriter:   os.Stderr,
		Exit:        os.Exit,
		uuid:        uuid.New(),
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
		OutWriter:   os.Stdout,
		ErrWriter:   os.Stderr,
		Exit:        os.Exit,
	}
}

// Context returns the underlying context.
func (r *RunContext) Context() context.Context {
	return r.ctx
}

// UUID returns the underlying run uuid. This can be used to globally identify the run context.
func (r *RunContext) UUID() uuid.UUID {
	return r.uuid
}

func (r *RunContext) SetContextValue(key string, value interface{}) {
	r.contextVals[key] = value
}

func (r *RunContext) ContextValues() map[string]interface{} {
	return r.contextVals
}

func (r *RunContext) EventEnv() map[string]interface{} {
	return r.EventEnvWithProjectContexts([]*ProjectContext{})
}

func (r *RunContext) EventEnvWithProjectContexts(projectContexts []*ProjectContext) map[string]interface{} {
	env := r.contextVals
	env["installId"] = r.State.InstallID

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

func (r *RunContext) loadInitialContextValues() {
	r.SetContextValue("version", baseVersion(version.Version))
	r.SetContextValue("fullVersion", version.Version)
	r.SetContextValue("isTest", IsTest())
	r.SetContextValue("isDev", IsDev())
	r.SetContextValue("os", runtime.GOOS)
	r.SetContextValue("ciPlatform", ciPlatform())
	r.SetContextValue("ciScript", ciScript())
	r.SetContextValue("ciPostCondition", os.Getenv("INFRACOST_CI_POST_CONDITION"))
	r.SetContextValue("ciPercentageThreshold", os.Getenv("INFRACOST_CI_PERCENTAGE_THRESHOLD"))
}

func (r *RunContext) IsCIRun() bool {
	return r.contextVals["ciPlatform"] != "" && !IsTest()
}

func baseVersion(v string) string {
	return strings.SplitN(v, "+", 2)[0]
}

func ciScript() string {
	if IsEnvPresent("INFRACOST_CI_IMAGE") {
		return "ci-image"
	} else if IsEnvPresent("INFRACOST_GITHUB_ACTION") {
		return "infracost-github-action"
	} else if IsEnvPresent("INFRACOST_CI_DIFF") {
		return "ci-diff"
	} else if IsEnvPresent("INFRACOST_CI_ATLANTIS_DIFF") {
		return "ci-atlantis-diff"
	} else if IsEnvPresent("INFRACOST_CI_JENKINS_DIFF") {
		return "ci-jenkins-diff"
	}

	return ""
}

// ciEnvMap contains information to detect what ci system the current process is running in.
type ciEnvMap struct {
	// vars is a list of OS env var names mapping to know ci system. If the key of this map is
	// present in OS env then it is assumed we are running in that CI system.
	vars map[string]string
	// prefixes contains a list of OS env var name prefixes. If the key of this map is
	// matches the prefix of any OS env then it is assumed we are running in that CI system.
	prefixes map[string]string
}

var ciMap = ciEnvMap{
	vars: map[string]string{
		"GITHUB_ACTIONS":       "github_actions",
		"GITLAB_CI":            "gitlab_ci",
		"CIRCLECI":             "circleci",
		"JENKINS_HOME":         "jenkins",
		"BUILDKITE":            "buildkite",
		"SYSTEM_COLLECTIONURI": fmt.Sprintf("azure_devops_%s", os.Getenv("BUILD_REPOSITORY_PROVIDER")),
		"TFC_RUN_ID":           "tfc",
		"ENV0_ENVIRONMENT_ID":  "env0",
		"SCALR_RUN_ID":         "scalr",
		"CF_BUILD_ID":          "codefresh",
		"TRAVIS":               "travis",
		"CODEBUILD_CI":         "codebuild",
		"TEAMCITY_VERSION":     "teamcity",
		"BUDDYBUILD_BRANCH":    "buddybuild",
		"BITRISE_IO":           "bitrise",
		"SEMAPHORE":            "semaphoreci",
		"APPVEYOR":             "appveyor",
		"WERCKER_GIT_BRANCH":   "wercker",
		"MAGNUM":               "magnumci",
		"SHIPPABLE":            "shippable",
		"TDDIUM":               "tddium",
		"GREENHOUSE":           "greenhouse",
		"CIRRUS_CI":            "cirrusci",
	},
	prefixes: map[string]string{
		"ATLANTIS_":  "atlantis",
		"BITBUCKET_": "bitbucket",
		"CONCOURSE_": "concourse",
		"SPACELIFT_": "spacelift",
		"HARNESS_":   "harness",
	},
}

func ciPlatform() string {
	for env, name := range ciMap.vars {
		if IsEnvPresent(env) {
			return name
		}
	}

	for _, k := range os.Environ() {
		for prefix, name := range ciMap.prefixes {
			if strings.HasPrefix(k, prefix) {
				return name
			}
		}
	}

	if IsEnvPresent("CI") {
		return "ci"
	}

	return ""
}

func ciVCSRepo() string {
	if IsEnvPresent("GITHUB_REPOSITORY") {
		serverURL := os.Getenv("GITHUB_SERVER_URL")
		if serverURL == "" {
			serverURL = "https://github.com"
		}
		return fmt.Sprintf("%s/%s", serverURL, os.Getenv("GITHUB_REPOSITORY"))
	} else if IsEnvPresent("CI_PROJECT_URL") {
		return os.Getenv("CI_PROJECT_URL")
	} else if IsEnvPresent("BUILD_REPOSITORY_URI") {
		return os.Getenv("BUILD_REPOSITORY_URI")
	} else if IsEnvPresent("BITBUCKET_GIT_HTTP_ORIGIN") {
		return os.Getenv("BITBUCKET_GIT_HTTP_ORIGIN")
	} else if IsEnvPresent("CIRCLE_REPOSITORY_URL") {
		return os.Getenv("CIRCLE_REPOSITORY_URL")
	}

	return ""
}

func ciVCSPullRequestURL() string {
	if IsEnvPresent("GITHUB_EVENT_PATH") && os.Getenv("GITHUB_EVENT_NAME") == "pull_request" {
		b, err := os.ReadFile(os.Getenv("GITHUB_EVENT_PATH"))
		if err != nil {
			log.Debugf("Error reading GITHUB_EVENT_PATH file: %v", err)
		}

		var event struct {
			PullRequest struct {
				HTMLURL string `json:"html_url"`
			} `json:"pull_request"`
		}

		err = json.Unmarshal(b, &event)
		if err != nil {
			log.Debugf("Error reading GITHUB_EVENT_PATH JSON: %v", err)
		}

		return event.PullRequest.HTMLURL
	} else if IsEnvPresent("CI_PROJECT_URL") && IsEnvPresent("CI_MERGE_REQUEST_IID") {
		return fmt.Sprintf("%s/merge_requests/%s", os.Getenv("CI_PROJECT_URL"), os.Getenv("CI_MERGE_REQUEST_IID"))
	}
	return ""
}
