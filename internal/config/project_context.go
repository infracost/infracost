package config

type ProjectContexter interface {
	ProjectContext() map[string]interface{}
}

type ProjectContext struct {
	Config        *Config
	ProjectConfig *Project
	runCtx        *RunContext
	contextVals   map[string]interface{}
}

func NewProjectContext(ctx *RunContext, projectCfg *Project) *ProjectContext {
	return &ProjectContext{
		Config:        ctx.Config,
		ProjectConfig: projectCfg,
		runCtx:        ctx,
		contextVals:   map[string]interface{}{},
	}
}

func EmptyProjectContext() *ProjectContext {
	return &ProjectContext{
		Config:        &Config{},
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

func (c *ProjectContext) SetFrom(d ProjectContexter) {
	m := d.ProjectContext()
	for k, v := range m {
		c.SetContextValue(k, v)
	}
}
