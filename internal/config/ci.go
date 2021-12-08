package config

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

func ciScript() string {
	if IsEnvPresent("INFRACOST_GITHUB_ACTION") {
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

func CIPlatform() string {
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
	} else if IsEnvPresent("CF_BUILD_ID") {
		return "codefresh"
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
