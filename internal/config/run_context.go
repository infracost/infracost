package config

import (
	"context"
	"fmt"
	"io"
	"maps"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/infracost/infracost/internal/logging"
	intSync "github.com/infracost/infracost/internal/sync"
	"github.com/infracost/infracost/internal/vcs"
	"github.com/infracost/infracost/internal/version"

	"github.com/google/uuid"
)

// ContextValues is a type that wraps a map with methods that safely
// handle concurrent reads and writes.
type ContextValues struct {
	values map[string]any
	mu     *sync.RWMutex
}

// NewContextValues returns a new instance of ContextValues.
func NewContextValues(values map[string]any) *ContextValues {
	return &ContextValues{
		values: values,
		mu:     &sync.RWMutex{},
	}
}

// GetValue safely retrieves a value from the map.
// It locks the mutex for reading, deferring the unlock until the method returns.
// This prevents a race condition if a separate goroutine writes to the map concurrently.
func (cv *ContextValues) GetValue(key string) (any, bool) {
	cv.mu.RLock()
	defer cv.mu.RUnlock()

	value, exists := cv.values[key]
	return value, exists
}

// SetValue safely sets a value in the map.
// It locks the mutex for writing, deferring the unlock until the method returns.
// This prevents a race condition if separate goroutines read or write to the map concurrently.
func (cv *ContextValues) SetValue(key string, value any) {
	cv.mu.Lock()
	defer cv.mu.Unlock()
	cv.values[key] = value
}

// Values safely retrieves a copy of the map.
// This method locks the mutex for reading, deferring the unlock until the method returns.
// By returning a copy, this prevents a race condition if separate goroutines read or write to the original map concurrently.
// However, creating a copy may be expensive for large maps.
func (cv *ContextValues) Values() map[string]any {
	cv.mu.RLock()
	defer cv.mu.RUnlock()
	copyMap := make(map[string]any)
	maps.Copy(copyMap, cv.values)
	return copyMap
}

type RunContext struct {
	ctx           context.Context
	uuid          uuid.UUID
	Config        *Config
	State         *State
	VCSMetadata   vcs.Metadata
	CMD           string
	ContextValues *ContextValues
	mu            *sync.RWMutex
	ModuleMutex   *intSync.KeyMutex
	StartTime     int64

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
		ctx:       rootCtx,
		OutWriter: os.Stdout,
		ErrWriter: os.Stderr,
		Exit:      os.Exit,
		uuid:      uuid.New(),
		Config:    cfg,
		State:     state,
		ContextValues: NewContextValues(
			map[string]any{
				"version":               baseVersion(version.Version),
				"fullVersion":           version.Version,
				"isTest":                IsTest(),
				"isDev":                 IsDev(),
				"os":                    runtime.GOOS,
				"ciPlatform":            ciPlatform(),
				"cliPlatform":           os.Getenv("INFRACOST_CLI_PLATFORM"),
				"ciScript":              ciScript(),
				"ciPostCondition":       os.Getenv("INFRACOST_CI_POST_CONDITION"),
				"ciPercentageThreshold": os.Getenv("INFRACOST_CI_PERCENTAGE_THRESHOLD"),
			}),
		ModuleMutex: &intSync.KeyMutex{},
		StartTime:   time.Now().Unix(),
	}

	return c, nil
}

func EmptyRunContext() *RunContext {
	return &RunContext{
		Config:        &Config{},
		State:         &State{},
		ContextValues: NewContextValues(map[string]any{}),
		mu:            &sync.RWMutex{},
		ModuleMutex:   &intSync.KeyMutex{},
		StartTime:     time.Now().Unix(),
		OutWriter:     os.Stdout,
		ErrWriter:     os.Stderr,
		Exit:          os.Exit,
	}
}

// IsAutoDetect returns true if the command is running with auto-detect functionality.
func (r *RunContext) IsAutoDetect() bool {
	return len(r.Config.Projects) <= 1 && r.Config.ConfigFilePath == ""
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

func (r *RunContext) EventEnv() map[string]any {
	return r.EventEnvWithProjectContexts([]*ProjectContext{})
}

func (r *RunContext) EventEnvWithProjectContexts(projectContexts []*ProjectContext) map[string]any {
	env := r.ContextValues.Values()

	env["installId"] = r.State.InstallID

	for _, projectContext := range projectContexts {
		if projectContext == nil {
			continue
		}

		for k, v := range projectContext.ContextValues.Values() {
			if _, ok := env[k]; !ok {
				env[k] = make([]any, 0)
			}
			env[k] = append(env[k].([]any), v)
		}
	}

	return env
}

func (r *RunContext) IsCIRun() bool {
	p, _ := r.ContextValues.GetValue("ciPlatform")
	return p != "" && !IsTest()
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
		logging.Logger.Debug().Str("is_cloud_enabled", fmt.Sprintf("%v", *r.Config.EnableCloud)).Msg("IsCloudEnabled explicitly set through Config.EnabledCloud")
		return *r.Config.EnableCloud
	}

	if r.Config.EnableCloudForOrganization {
		logging.Logger.Debug().Msg("IsCloudEnabled is true with org level setting enabled.")
		return true
	}

	logging.Logger.Debug().Str("is_cloud_enabled", fmt.Sprintf("%v", r.Config.EnableDashboard)).Msg("IsCloudEnabled inferred from Config.EnabledDashboard")
	return r.Config.EnableDashboard
}

func (r *RunContext) IsCloudUploadEnabled() bool {
	if r.Config.EnableCloudUpload != nil {
		return *r.Config.EnableCloudUpload
	}
	return r.IsCloudEnabled()
}

// IsCloudUploadExplicitlyEnabled returns true if cloud upload has been enabled through one of the
// env variables ENABLE_CLOUD, ENABLE_CLOUD_UPLOAD, or ENABLE_DASHBOARD
func (r *RunContext) IsCloudUploadExplicitlyEnabled() bool {
	return r.IsCloudUploadEnabled() &&
		(r.Config.EnableCloud != nil || r.Config.EnableCloudUpload != nil || r.Config.EnableDashboard)
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
		"ATLANTIS_":       "atlantis",
		"BITBUCKET_":      "bitbucket",
		"CONCOURSE_":      "concourse",
		"SPACELIFT_":      "spacelift",
		"HARNESS_":        "harness",
		"TERRATEAM_":      "terrateam",
		"KEPTN_":          "keptn",
		"CLOUDCONCIERGE_": "cloudconcierge",
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
