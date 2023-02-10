package config

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/infracost/infracost/internal/logging"
	intSync "github.com/infracost/infracost/internal/sync"
	"github.com/infracost/infracost/internal/vcs"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/infracost/infracost/internal/ui"
	"github.com/infracost/infracost/internal/version"
)

type RunContext struct {
	ctx         context.Context
	uuid        uuid.UUID
	Config      *Config
	State       *State
	VCSMetadata vcs.Metadata
	CMD         string
	contextVals map[string]interface{}
	mu          *sync.RWMutex
	ModuleMutex *intSync.KeyMutex
	StartTime   int64

	isCommentCmd bool

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
		mu:          &sync.RWMutex{},
		ModuleMutex: &intSync.KeyMutex{},
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
		mu:          &sync.RWMutex{},
		ModuleMutex: &intSync.KeyMutex{},
		StartTime:   time.Now().Unix(),
		OutWriter:   os.Stdout,
		ErrWriter:   os.Stderr,
		Exit:        os.Exit,
	}
}

var (
	outputIndent = "  "
)

// NewSpinner returns an ui.Spinner built from the RunContext.
func (r *RunContext) NewSpinner(msg string) *ui.Spinner {
	return ui.NewSpinner(msg, ui.SpinnerOptions{
		EnableLogging: r.Config.IsLogging(),
		NoColor:       r.Config.NoColor,
		Indent:        outputIndent,
	})
}

func (r *RunContext) GetParallelism() (int, error) {
	var parallelism int

	if r.Config.Parallelism == nil {
		parallelism = 4
		numCPU := runtime.NumCPU()
		if numCPU*4 > parallelism {
			parallelism = numCPU * 4
		}

		if parallelism > 16 {
			parallelism = 16
		}

		return parallelism, nil
	}

	parallelism = *r.Config.Parallelism

	if parallelism < 0 {
		return parallelism, fmt.Errorf("parallelism must be a positive number")
	}

	if parallelism > 16 {
		return parallelism, fmt.Errorf("parallelism must be less than 16")
	}

	return parallelism, nil
}

// Context returns the underlying context.
func (r *RunContext) Context() context.Context {
	return r.ctx
}

// UUID returns the underlying run uuid. This can be used to globally identify the run context.
func (r *RunContext) UUID() uuid.UUID {
	return r.uuid
}

func (r *RunContext) VCSRepositoryURL() string {
	return r.VCSMetadata.Remote.URL
}

func (r *RunContext) SetContextValue(key string, value interface{}) {
	r.mu.Lock()
	r.contextVals[key] = value
	r.mu.Unlock()
}

func (r *RunContext) ContextValues() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.contextVals
}

func (r *RunContext) GetResourceWarnings() map[string]map[string]int {
	contextValues := r.ContextValues()

	if warnings := contextValues["resourceWarnings"]; warnings != nil {
		return warnings.(map[string]map[string]int)
	}

	return nil
}

func (r *RunContext) SetResourceWarnings(resourceWarnings map[string]map[string]int) {
	r.SetContextValue("resourceWarnings", resourceWarnings)
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
	r.SetContextValue("cliPlatform", os.Getenv("INFRACOST_CLI_PLATFORM"))
	r.SetContextValue("ciScript", ciScript())
	r.SetContextValue("ciPostCondition", os.Getenv("INFRACOST_CI_POST_CONDITION"))
	r.SetContextValue("ciPercentageThreshold", os.Getenv("INFRACOST_CI_PERCENTAGE_THRESHOLD"))
}

func (r *RunContext) IsCIRun() bool {
	return r.contextVals["ciPlatform"] != "" && !IsTest()
}

// SetIsInfracostComment identifies that the primary command being run is `infracost comment`
func (r *RunContext) SetIsInfracostComment() {
	r.isCommentCmd = true
}

func (r *RunContext) IsInfracostComment() bool {
	return r.isCommentCmd
}

func (r *RunContext) IsCloudEnabled() bool {
	if r.Config.EnableCloud != nil {
		logging.Logger.WithFields(log.Fields{"is_cloud_enabled": *r.Config.EnableCloud}).Debug("IsCloudEnabled explicitly set through Config.EnabledCloud")
		return *r.Config.EnableCloud
	}

	if r.Config.EnableCloudForOrganization {
		logging.Logger.Debug("IsCloudEnabled is true with org level setting enabled.")
		return true
	}

	logging.Logger.WithFields(log.Fields{"is_cloud_enabled": r.Config.EnableDashboard}).Debug("IsCloudEnabled inferred from Config.EnabledDashboard")
	return r.Config.EnableDashboard
}

func (r *RunContext) IsCloudUploadEnabled() bool {
	if r.Config.EnableCloudUpload != nil {
		return *r.Config.EnableCloudUpload
	}
	return r.IsCloudEnabled()
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
		"TS_ENV":               "terraspace",
	},
	prefixes: map[string]string{
		"ATLANTIS_":  "atlantis",
		"BITBUCKET_": "bitbucket",
		"CONCOURSE_": "concourse",
		"SPACELIFT_": "spacelift",
		"HARNESS_":   "harness",
		"TERRATEAM_": "terrateam",
		"KEPTN_":     "keptn",
	},
}

func ciPlatform() string {
	if os.Getenv("INFRACOST_CI_PLATFORM") != "" {
		return os.Getenv("INFRACOST_CI_PLATFORM")
	}

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
