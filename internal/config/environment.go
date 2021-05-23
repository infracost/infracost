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

type Environment struct {
	Version                             string   `json:"version"`
	FullVersion                         string   `json:"fullVersion"`
	IsTest                              bool     `json:"isTest"`
	IsDev                               bool     `json:"isDev"`
	InstallID                           string   `json:"installId"`
	IsDefaultPricingAPIEndpoint         bool     `json:"isDefaultPricingAPIEndpoint"`
	OS                                  string   `json:"os"`
	CIPlatform                          string   `json:"ciPlatform,omitempty"`
	CIScript                            string   `json:"ciScript,omitempty"`
	Command                             string   `json:"command"`
	Flags                               []string `json:"flags"`
	OutputFormat                        string   `json:"outputFormat"`
	ProjectType                         string   `json:"projectType"`
	TerraformBinary                     string   `json:"terraformBinary"`
	TerraformFullVersion                string   `json:"terraformFullVersion"`
	TerraformVersion                    string   `json:"terraformVersion"`
	TerraformRemoteExecutionModeEnabled bool     `json:"terraformRemoteExecutionModeEnabled"`
	TerraformInfracostProviderEnabled   bool     `json:"terraformInfracostProviderEnabled"`
	IsAWSChina                          bool     `json:"isAwsChina"`
	HasConfigFile                       bool     `json:"hasConfigFile"`
	HasUsageFile                        bool     `json:"hasUsageFile"`
}

func NewEnvironment() *Environment {
	return &Environment{
		Version:     baseVersion(version.Version),
		FullVersion: version.Version,
		IsTest:      isTest(),
		IsDev:       isDev(),
		OS:          runtime.GOOS,
		CIPlatform:  ciPlatform(),
		CIScript:    ciScript(),
	}
}

func (e *Environment) SetProjectEnvironment(projectType string, projectCfg *Project) {
	e.ProjectType = projectType

	if projectType == "terraform_dir" || projectType == "terraform_plan" {
		binary := projectCfg.TerraformBinary
		if binary == "" {
			binary = "terraform"
		}
		out, _ := exec.Command(binary, "-version").Output()
		fullVersion := strings.SplitN(string(out), "\n", 2)[0]
		version := terraformVersion(fullVersion)

		e.TerraformBinary = binary
		e.TerraformFullVersion = fullVersion
		e.TerraformVersion = version
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
	} else if IsTruthy(os.Getenv("INFRACOST_CI_JENKINS_DIFF")) {
		return "ci-jenkins-diff"
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
	} else if IsTruthy(os.Getenv("JENKINS_HOME")) {
		return "jenkins"
	} else if IsTruthy(os.Getenv("SYSTEM_COLLECTIONURI")) {
		return fmt.Sprintf("azure_devops_%s", os.Getenv("BUILD_REPOSITORY_PROVIDER"))
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

func AddAuthHeaders(apiKey string, req *http.Request) {
	AddNoAuthHeaders(req)
	req.Header.Set("X-Api-Key", apiKey)
	req.Header.Set("X-Trace-Id", TraceID())
}
