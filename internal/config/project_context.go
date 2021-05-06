package config

import (
	"os/exec"
	"strings"
)

type ProjectMetadata struct {
	ProjectType                         string `json:"projectType"`
	TerraformBinary                     string `json:"terraformBinary"`
	TerraformFullVersion                string `json:"terraformFullVersion"`
	TerraformVersion                    string `json:"terraformVersion"`
	TerraformRemoteExecutionModeEnabled bool   `json:"terraformRemoteExecutionModeEnabled"`
	TerraformInfracostProviderEnabled   bool   `json:"terraformInfracostProviderEnabled"`
	IsAWSChina                          bool   `json:"isAwsChina"`
	HasUsageFile                        bool   `json:"hasUsageFile"`
}

type ProjectContext struct {
	RunContext    *RunContext
	ProjectConfig *Project
	Metadata      ProjectMetadata
}

func NewProjectContext(runCtx *RunContext, projectCfg *Project) *ProjectContext {
	return &ProjectContext{
		RunContext:    runCtx,
		ProjectConfig: projectCfg,
		Metadata:      ProjectMetadata{},
	}
}

func EmptyProjectContext() *ProjectContext {
	return &ProjectContext{
		RunContext:    EmptyRunContext(),
		ProjectConfig: &Project{},
		Metadata:      ProjectMetadata{},
	}
}

// TODO: Include run metadata
func (c *ProjectContext) AllMetadata() ProjectMetadata {
	return c.Metadata
}

func (c *ProjectContext) LoadMetadataForProjectType(projectType string) {
	c.Metadata.ProjectType = projectType

	if projectType == "terraform_dir" || projectType == "terraform_plan" {
		binary := c.ProjectConfig.TerraformBinary
		if binary == "" {
			binary = "terraform"
		}
		out, _ := exec.Command(binary, "-version").Output()
		fullVersion := strings.SplitN(string(out), "\n", 2)[0]
		version := terraformVersion(fullVersion)

		c.Metadata.TerraformBinary = binary
		c.Metadata.TerraformFullVersion = fullVersion
		c.Metadata.TerraformVersion = version
	}
}
