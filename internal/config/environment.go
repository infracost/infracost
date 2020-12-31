package config

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/infracost/infracost/internal/version"
)

type EnvironmentSpec struct {
	Version                             string   `json:"version"`
	FullVersion                         string   `json:"fullVersion"`
	Flags                               []string `json:"flags"`
	OutputFormat                        string   `json:"outputFormat"`
	IsTest                              bool     `json:"isTest"`
	IsDev                               bool     `json:"isDev"`
	OS                                  string   `json:"os"`
	CIPlatform                          string   `json:"ciPlatform,omitempty"`
	CIScript                            string   `json:"ciScript,omitempty"`
	TerraformBinary                     string   `json:"terraformBinary"`
	TerraformFullVersion                string   `json:"terraformFullVersion"`
	TerraformVersion                    string   `json:"terraformVersion"`
	TerraformRemoteExecutionModeEnabled bool     `json:"terraformRemoteExecutionModeEnabled"`
	TerraformInfracostProviderEnabled   bool     `json:"terraformInfracostProviderEnabled"`
}

var Environment *EnvironmentSpec

func init() {
	Environment = loadEnvironment()
}

func loadEnvironment() *EnvironmentSpec {
	return &EnvironmentSpec{
		Version:                             baseVersion(version.Version),
		FullVersion:                         version.Version,
		Flags:                               []string{},
		OutputFormat:                        "",
		IsTest:                              isTest(),
		IsDev:                               isDev(),
		OS:                                  runtime.GOOS,
		CIPlatform:                          ciPlatform(),
		CIScript:                            ciScript(),
		TerraformBinary:                     filepath.Base(terraformBinary()),
		TerraformFullVersion:                terraformFullVersion(),
		TerraformVersion:                    terraformVersion(),
		TerraformRemoteExecutionModeEnabled: false,
		TerraformInfracostProviderEnabled:   false,
	}
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

func terraformBinary() string {
	terraformBinary := os.Getenv("TERRAFORM_BINARY")
	if terraformBinary == "" {
		terraformBinary = "terraform"
	}
	return terraformBinary
}

func terraformFullVersion() string {
	exe := terraformBinary()
	out, _ := exec.Command(exe, "-version").Output()

	return strings.SplitN(string(out), "\n", 2)[0]
}

func terraformVersion() string {
	v := terraformFullVersion()

	p := strings.Split(v, " ")
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
