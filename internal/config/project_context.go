package config

import (
	"sync"
)

type ProjectContexter interface {
	ProjectContext() map[string]interface{}
}

type ProjectContext struct {
	RunContext    *RunContext
	ProjectConfig *Project
	ContextValues *ContextValues
	mu            *sync.RWMutex

	UsingCache bool
	CacheErr   string
}

func NewProjectContext(runCtx *RunContext, projectCfg *Project) *ProjectContext {
	return &ProjectContext{
		RunContext:    runCtx,
		ProjectConfig: projectCfg,
		ContextValues: NewContextValues(map[string]interface{}{}),
		mu:            &sync.RWMutex{},
	}
}

func (c *ProjectContext) SetProjectType(projectType string) {
	c.ContextValues.SetValue("project_type", projectType)
	var projectTypes []interface{}
	if t, ok := c.RunContext.ContextValues.GetValue("projectTypes"); ok {
		projectTypes = t.([]interface{})
	}

	projectTypes = append(projectTypes, projectType)
	c.RunContext.ContextValues.SetValue("projectTypes", projectTypes)
}

func (c *ProjectContext) SetFrom(d ProjectContexter) {
	m := d.ProjectContext()
	for k, v := range m {
		c.ContextValues.SetValue(k, v)
	}
}
