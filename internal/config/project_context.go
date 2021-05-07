package config

type ProjectContext struct {
	RunContext    *RunContext
	ProjectConfig *Project
	metadata      map[string]interface{}
}

func NewProjectContext(runCtx *RunContext, projectCfg *Project) *ProjectContext {
	return &ProjectContext{
		RunContext:    runCtx,
		ProjectConfig: projectCfg,
		metadata:      map[string]interface{}{},
	}
}

func EmptyProjectContext() *ProjectContext {
	return &ProjectContext{
		RunContext:    EmptyRunContext(),
		ProjectConfig: &Project{},
		metadata:      map[string]interface{}{},
	}
}

func (c *ProjectContext) SetMetadata(key string, value interface{}) {
	c.metadata[key] = value
}

func (c *ProjectContext) AllMetadata() map[string]interface{} {
	return map[string]interface{}{
		"run":     c.RunContext.metadata,
		"project": c.metadata,
	}
}
