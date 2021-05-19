package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	log "github.com/sirupsen/logrus"
)

type ProjectContext struct {
	RunContext    *RunContext
	ProjectConfig *Project
	contextVals   map[string]interface{}
}

func NewProjectContext(runCtx *RunContext, projectCfg *Project) *ProjectContext {
	return &ProjectContext{
		RunContext:    runCtx,
		ProjectConfig: projectCfg,
		contextVals:   map[string]interface{}{},
	}
}

func EmptyProjectContext() *ProjectContext {
	return &ProjectContext{
		RunContext:    EmptyRunContext(),
		ProjectConfig: &Project{},
		contextVals:   map[string]interface{}{},
	}
}

func (c *ProjectContext) SetContextValue(key string, value interface{}) {
	c.contextVals[key] = value
}

func (c *ProjectContext) ContextValues() map[string]interface{} {
	return c.contextVals
}

func DetectProjectMetadata(ctx *ProjectContext) *schema.ProjectMetadata {
	vcsRepoURL := os.Getenv("INFRACOST_VCS_REPOSITORY_URL")
	vcsSubPath := os.Getenv("INFRACOST_VCS_SUB_PATH")
	vcsPullRequestURL := os.Getenv("INFRACOST_VCS_PULL_REQUEST_URL")

	if vcsRepoURL == "" {
		vcsRepoURL = gitRepo(ctx.ProjectConfig.Path)
	}

	if vcsRepoURL != "" && vcsSubPath == "" {
		vcsSubPath = gitSubPath(ctx.ProjectConfig.Path)
	}

	return &schema.ProjectMetadata{
		Path:              ctx.ProjectConfig.Path,
		VCSRepoURL:        vcsRepoURL,
		VCSSubPath:        vcsSubPath,
		VCSPullRequestURL: vcsPullRequestURL,
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
