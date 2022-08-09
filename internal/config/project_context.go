package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/sirupsen/logrus"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/vcs"
)

type ProjectContexter interface {
	ProjectContext() map[string]interface{}
}

type ProjectContext struct {
	RunContext    *RunContext
	ProjectConfig *Project
	logger        *logrus.Entry
	contextVals   map[string]interface{}
	mu            *sync.RWMutex

	UsingCache bool
	CacheErr   string
}

func NewProjectContext(runCtx *RunContext, projectCfg *Project, fields logrus.Fields) *ProjectContext {
	contextLogger := logging.Logger.WithFields(fields)

	return &ProjectContext{
		RunContext:    runCtx,
		ProjectConfig: projectCfg,
		logger:        contextLogger,
		contextVals:   map[string]interface{}{},
		mu:            &sync.RWMutex{},
	}
}

func (p *ProjectContext) Logger() *logrus.Entry {
	if p.logger == nil {
		return logging.Logger.WithFields(p.logFields())
	}

	return p.logger.WithFields(p.logFields())
}

func (p *ProjectContext) logFields() logrus.Fields {
	return logrus.Fields{
		"project_name": p.ProjectConfig.Name,
		"project_path": p.ProjectConfig.Path,
	}
}

func (c *ProjectContext) SetContextValue(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.contextVals[key] = value
}

func (c *ProjectContext) ContextValues() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
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

	meta, err := vcs.MetadataFetcher.Get(path)
	if err != nil {
		logging.Logger.WithError(err).Debugf("failed to fetch vcs metadata for path %s", path)
	}

	if vcsPullRequestURL == "" && meta.PullRequest != nil {
		vcsPullRequestURL = meta.PullRequest.URL
	}

	pm := &schema.ProjectMetadata{
		Path:               path,
		TerraformWorkspace: terraformWorkspace,
		Branch:             meta.Branch.Name,
		Commit:             meta.Commit.SHA,
		CommitAuthorEmail:  meta.Commit.AuthorEmail,
		CommitAuthorName:   meta.Commit.AuthorName,
		CommitTimestamp:    meta.Commit.Time.UTC(),
		CommitMessage:      meta.Commit.Message,
		VCSRepoURL:         vcsRepoURL,
		VCSSubPath:         vcsSubPath,
		VCSPullRequestURL:  vcsPullRequestURL,
	}

	if meta.PullRequest != nil {
		pm.VCSProvider = meta.PullRequest.VCSProvider
		pm.VCSPullRequestID = meta.PullRequest.ID
		pm.VCSBaseBranch = meta.PullRequest.BaseBranch
		pm.VCSPullRequestTitle = meta.PullRequest.Title
		pm.VCSPullRequestAuthor = meta.PullRequest.Author
	}

	if meta.Pipeline != nil {
		pm.VCSPipelineRunID = meta.Pipeline.ID
	}

	return pm
}

func gitRepo(path string) string {
	logging.Logger.Debugf("Checking if %s is a git repo", path)
	cmd := exec.Command("git", "ls-remote", "--get-url")

	if isDir(path) {
		cmd.Dir = path
	} else {
		cmd.Dir = filepath.Dir(path)
	}

	out, err := cmd.Output()
	if err != nil {
		logging.Logger.WithError(err).Debugf("could not detect a git repo at %s", path)
		return ""
	}
	return strings.Split(string(out), "\n")[0]
}

func gitSubPath(path string) string {
	topLevel, err := gitToplevel(path)
	if err != nil {
		logging.Logger.WithError(err).Debugf("Could not get git top level directory for %s", path)
		return ""
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		logging.Logger.WithError(err).Debugf("Could not get absolute path for %s", path)
		return ""
	}

	subPath, err := filepath.Rel(topLevel, absPath)
	if err != nil {
		logging.Logger.WithError(err).Debugf("Could not get relative path for %s from %s", absPath, topLevel)
		return ""
	}

	if subPath == "." {
		return ""
	}

	return subPath
}

func gitToplevel(path string) (string, error) {
	r, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return "", fmt.Errorf("failed to detect a git directory in path %s of any of its parent dirs %w", path, err)
	}
	wt, err := r.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to return worktree for path %s %w", path, err)
	}

	return wt.Filesystem.Root(), nil
}

func stripVCSRepoPassword(repoURL string) string {
	r := regexp.MustCompile(`.*:([^@]*)@`)
	return r.ReplaceAllString(repoURL, "")
}
