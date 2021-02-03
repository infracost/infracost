package config

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/infracost/infracost/internal/version"
)

type EnvironmentSpec struct {
	Version                             string   `json:"version"`
	FullVersion                         string   `json:"fullVersion"`
	IsTest                              bool     `json:"isTest"`
	IsDev                               bool     `json:"isDev"`
	OS                                  string   `json:"os"`
	CIPlatform                          string   `json:"ciPlatform,omitempty"`
	CIScript                            string   `json:"ciScript,omitempty"`
	Flags                               []string `json:"flags"`
	OutputFormat                        string   `json:"outputFormat"`
	TerraformBinary                     string   `json:"terraformBinary"`
	TerraformFullVersion                string   `json:"terraformFullVersion"`
	TerraformVersion                    string   `json:"terraformVersion"`
	TerraformRemoteExecutionModeEnabled bool     `json:"terraformRemoteExecutionModeEnabled"`
	TerraformInfracostProviderEnabled   bool     `json:"terraformInfracostProviderEnabled"`
	IsAWSChina                          bool     `json:"isAwsChina"`
	HasConfigFile                       bool     `json:"HasConfigFile"`
	HasUsageFile                        bool     `json:"hasUsageFile"`
}

var Environment *EnvironmentSpec

func init() {
	loadInitialEnvironment()
}

func loadInitialEnvironment() {
	Environment = &EnvironmentSpec{
		Version:     baseVersion(version.Version),
		FullVersion: version.Version,
		IsTest:      isTest(),
		IsDev:       isDev(),
		OS:          runtime.GOOS,
		CIPlatform:  ciPlatform(),
		CIScript:    ciScript(),
	}
}

func LoadTerraformEnvironment(projectConfig *TerraformProjectSpec) {
	binary := projectConfig.Binary
	if binary == "" {
		binary = "terraform"
	}
	out, _ := exec.Command(binary, "-version").Output()
	fullVersion := strings.SplitN(string(out), "\n", 2)[0]
	version := terraformVersion(fullVersion)

	Environment.TerraformBinary = binary
	Environment.TerraformFullVersion = fullVersion
	Environment.TerraformVersion = version
}

func LoadOutputEnvironment(outputConfig *OutputSpec) {
	Environment.OutputFormat = outputConfig.Format
}

func userAgent() string {
	userAgent := "infracost"

	if version.Version != "" {
		userAgent += fmt.Sprintf("-%s", version.Version)
	}

	return userAgent
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

func TraceID() string {
	return uuid.New().String()
}

func AddNoAuthHeaders(req *http.Request) {
	req.Header.Set("content-type", "application/json")
	req.Header.Set("User-Agent", userAgent())
}

func AddAuthHeaders(req *http.Request) {
	AddNoAuthHeaders(req)
	req.Header.Set("X-Api-Key", Config.APIKey)
	req.Header.Set("X-Trace-Id", TraceID())
}
