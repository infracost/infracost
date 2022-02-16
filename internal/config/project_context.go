package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/infracost/infracost/internal/schema"
)

type ProjectContexter interface {
	ProjectContext() map[string]interface{}
}

type ProjectContext struct {
	RunContext    *RunContext
	ProjectConfig *Project
	contextVals   map[string]interface{}
	mu            *sync.Mutex

	UsingCache bool
	CacheErr   string
}

func NewProjectContext(runCtx *RunContext, projectCfg *Project) *ProjectContext {
	return &ProjectContext{
		RunContext:    runCtx,
		ProjectConfig: projectCfg,
		contextVals:   map[string]interface{}{},
		mu:            &sync.Mutex{},
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
	c.mu.Lock()
	defer c.mu.Unlock()
	c.contextVals[key] = value
}

func (c *ProjectContext) ContextValues() map[string]interface{} {
	return c.contextVals
}

func (c *ProjectContext) SetFrom(d ProjectContexter) {
	m := d.ProjectContext()
	for k, v := range m {
		c.SetContextValue(k, v)
	}
}

func DetectProjectMetadata(path string) *schema.ProjectMetadata {
	vcsRepoURL := os.Getenv("INFRACOST_VCS_REPOSITORY_URL")
	vcsSubPath := os.Getenv("INFRACOST_VCS_SUB_PATH")
	vcsPullRequestURL := os.Getenv("INFRACOST_VCS_PULL_REQUEST_URL")
	terraformWorkspace := os.Getenv("INFRACOST_TERRAFORM_WORKSPACE")

	if vcsRepoURL == "" {
		vcsRepoURL = ciVCSRepo()
	}

	if vcsRepoURL == "" {
		vcsRepoURL = gitRepo(path)
	}

	if vcsRepoURL != "" && vcsSubPath == "" {
		vcsSubPath = gitSubPath(path)
	}

	if vcsPullRequestURL == "" {
		vcsPullRequestURL = ciVCSPullRequestURL()
	}

	vcsRepoURL = stripVCSRepoPassword(vcsRepoURL)

	return &schema.ProjectMetadata{
		Path:               path,
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

func stripVCSRepoPassword(repoURL string) string {
	r := regexp.MustCompile(`.*:([^@]*)@`)
	return r.ReplaceAllString(repoURL, "")
}
