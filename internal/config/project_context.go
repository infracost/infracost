package config

import (
	"context"
	"sync"

	"github.com/rs/zerolog"

	"github.com/infracost/infracost/internal/logging"
)

type ProjectContexter interface {
	ProjectContext() map[string]any
}

type ProjectContext struct {
	RunContext    *RunContext
	ProjectConfig *Project
	logger        zerolog.Logger
	ContextValues *ContextValues
	mu            *sync.RWMutex

	UsingCache bool
	CacheErr   string
}

func NewProjectContext(runCtx *RunContext, projectCfg *Project, logFields any) *ProjectContext {
	ctx := logging.Logger.With().
		Str("project_name", projectCfg.Name).
		Str("project_path", projectCfg.Path)

	if logFields != nil {
		switch v := logFields.(type) {
		case context.Context:
			ctx = ctx.Ctx(v)
		default:
			ctx = ctx.Fields(v)
		}
	}

	contextLogger := ctx.Logger()

	return &ProjectContext{
		RunContext:    runCtx,
		ProjectConfig: projectCfg,
		logger:        contextLogger,
		ContextValues: NewContextValues(map[string]any{}),
		mu:            &sync.RWMutex{},
	}
}

func (c *ProjectContext) SetProjectType(projectType string) {
	c.ContextValues.SetValue("project_type", projectType)
	var projectTypes []any
	if t, ok := c.RunContext.ContextValues.GetValue("projectTypes"); ok {
		projectTypes = t.([]any)
	}

	projectTypes = append(projectTypes, projectType)
	c.RunContext.ContextValues.SetValue("projectTypes", projectTypes)
}

func (c *ProjectContext) Logger() zerolog.Logger {
	return c.logger
}

func (c *ProjectContext) SetFrom(d ProjectContexter) {
	m := d.ProjectContext()
	for k, v := range m {
		c.ContextValues.SetValue(k, v)
	}
}
