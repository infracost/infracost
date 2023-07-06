package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/sirupsen/logrus"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/schema"
)

type ProjectContexter interface {
	ProjectContext() map[string]interface{}
}

type ProjectContext struct {
	RunContext    *RunContext
	ProjectConfig *Project
	logger        *logrus.Entry
	ContextValues *ContextValues
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
		ContextValues: NewContextValues(map[string]interface{}{}),
		mu:            &sync.RWMutex{},
	}
}

func (c *ProjectContext) Logger() *logrus.Entry {
	if c.logger == nil {
		return logging.Logger.WithFields(c.logFields())
	}

	return c.logger.WithFields(c.logFields())
}

func (c *ProjectContext) logFields() logrus.Fields {
	return logrus.Fields{
		"project_name": c.ProjectConfig.Name,
		"project_path": c.ProjectConfig.Path,
	}
}

func (c *ProjectContext) SetFrom(d ProjectContexter) {
	m := d.ProjectContext()
	for k, v := range m {
		c.ContextValues.SetValue(k, v)
	}
}

func DetectProjectMetadata(path string) *schema.ProjectMetadata {
	vcsSubPath := os.Getenv("INFRACOST_VCS_SUB_PATH")
	terraformWorkspace := os.Getenv("INFRACOST_TERRAFORM_WORKSPACE")

	if vcsSubPath == "" {
		vcsSubPath = gitSubPath(path)
	}

	return &schema.ProjectMetadata{
		Path:               path,
		TerraformWorkspace: terraformWorkspace,
		VCSSubPath:         vcsSubPath,
	}
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
