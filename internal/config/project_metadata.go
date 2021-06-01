package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	log "github.com/sirupsen/logrus"
)

func DetectProjectMetadata(projectCfg *Project) *schema.ProjectMetadata {
	vcsRepoURL := os.Getenv("INFRACOST_VCS_REPOSITORY_URL")
	vcsSubPath := os.Getenv("INFRACOST_VCS_SUB_PATH")
	vcsPullRequestURL := os.Getenv("INFRACOST_VCS_PULL_REQUEST_URL")
	terraformWorkspace := os.Getenv("INFRACOST_TERRAFORM_WORKSPACE")

	if vcsRepoURL == "" {
		vcsRepoURL = gitRepo(projectCfg.Path)
	}

	if vcsRepoURL != "" && vcsSubPath == "" {
		vcsSubPath = gitSubPath(projectCfg.Path)
	}

	return &schema.ProjectMetadata{
		Path:               projectCfg.Path,
		VCSRepoURL:         vcsRepoURL,
		VCSSubPath:         vcsSubPath,
		VCSPullRequestURL:  vcsPullRequestURL,
		TerraformWorkspace: terraformWorkspace,
	}
}

func gitRepo(path string) string {
	log.Debugf("Checking if %s is a git repo", path)
	cmd := exec.Command("git", "ls-remote", "--get-url")

	if isDir(path) {
		cmd.Dir = path
	} else {
		cmd.Dir = filepath.Dir(path)
	}

	out, err := cmd.Output()
	if err != nil {
		log.Debugf("Could not detect a git repo at %s", path)
		return ""
	}
	return strings.Split(string(out), "\n")[0]
}

func gitSubPath(path string) string {
	topLevel, err := gitToplevel(path)
	if err != nil {
		log.Debugf("Could not get git top level directory for %s", path)
		return ""
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Debugf("Could not get absolute path for %s", path)
		return ""
	}

	subPath, err := filepath.Rel(topLevel, absPath)
	if err != nil {
		log.Debugf("Could not get relative path for %s from %s", absPath, topLevel)
		return ""
	}

	return subPath
}

func gitToplevel(path string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")

	if isDir(path) {
		cmd.Dir = path
	} else {
		cmd.Dir = filepath.Dir(path)
	}

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.Split(string(out), "\n")[0], nil
}
