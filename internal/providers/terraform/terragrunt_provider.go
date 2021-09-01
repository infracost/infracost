package terraform

import (
	"bytes"
	"encoding/json"
	"os"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/ui"
	"github.com/pkg/errors"
)

var defaultTerragruntBinary = "terragrunt"
var minTerragruntVer = "v0.28.1"

type TerragruntProvider struct {
	ctx  *config.ProjectContext
	Path string
	*DirProvider
}

type TerragruntInfo struct {
	WorkingDir string
}

func NewTerragruntProvider(ctx *config.ProjectContext) schema.Provider {
	dirProvider := NewDirProvider(ctx).(*DirProvider)

	terragruntBinary := ctx.ProjectConfig.TerraformBinary
	if terragruntBinary == "" {
		terragruntBinary = defaultTerragruntBinary
	}

	dirProvider.TerraformBinary = terragruntBinary
	dirProvider.IsTerragrunt = true

	return &TerragruntProvider{
		ctx:         ctx,
		DirProvider: dirProvider,
		Path:        ctx.ProjectConfig.Path,
	}
}

func (p *TerragruntProvider) Type() string {
	return "terragrunt"
}

func (p *TerragruntProvider) DisplayType() string {
	return "Terragrunt directory"
}

func (p *TerragruntProvider) AddMetadata(metadata *schema.ProjectMetadata) {
	// no op
}

func (p *TerragruntProvider) LoadResources(usage map[string]*schema.UsageData) ([]*schema.Project, error) {
	paths, err := p.getProjectPaths()
	if err != nil {
		return []*schema.Project{}, err
	}

	var outs [][]byte

	if p.UseState {
		outs, err = p.generateStateJSONs(paths)
	} else {
		outs, err = p.generatePlanJSONs(paths)
	}
	if err != nil {
		return []*schema.Project{}, err
	}

	projects := make([]*schema.Project, 0, len(paths))

	for i, path := range paths {
		metadata := config.DetectProjectMetadata(path)
		metadata.Type = p.Type()
		p.AddMetadata(metadata)
		name := schema.GenerateProjectName(metadata, p.ctx.RunContext.Config.EnableDashboard)

		project := schema.NewProject(name, metadata)

		parser := NewParser(p.ctx)
		pastResources, resources, err := parser.parseJSON(outs[i], usage)
		if err != nil {
			return projects, errors.Wrap(err, "Error parsing Terraform JSON")
		}

		project.HasDiff = !p.UseState
		if project.HasDiff {
			project.PastResources = pastResources
		}
		project.Resources = resources

		projects = append(projects, project)
	}

	return projects, nil
}

func (p *TerragruntProvider) getProjectPaths() ([]string, error) {
	opts := &CmdOptions{
		TerraformBinary: p.TerraformBinary,
		Dir:             p.Path,
	}
	out, err := Cmd(opts, "run-all", "--terragrunt-ignore-external-dependencies", "terragrunt-info")
	if err != nil {
		return []string{}, err
	}

	jsons := bytes.SplitAfter(out, []byte{'}', '\n'})
	if len(jsons) > 1 {
		jsons = jsons[:len(jsons)-1]
	}

	paths := make([]string, 0, len(jsons))
	for _, j := range jsons {
		var info TerragruntInfo
		err = json.Unmarshal(j, &info)
		if err != nil {
			return paths, err
		}

		paths = append(paths, info.WorkingDir)
	}

	return paths, nil
}

func (p *TerragruntProvider) generateStateJSONs(paths []string) ([][]byte, error) {
	err := p.checks()
	if err != nil {
		return [][]byte{}, err
	}

	outs := make([][]byte, 0, len(paths))

	spinnerMsg := "Running terragrunt show"
	if len(paths) > 1 {
		spinnerMsg += " for each project"
	}
	spinner := ui.NewSpinner(spinnerMsg, p.spinnerOpts)

	for _, path := range paths {
		opts, err := p.buildCommandOpts(path)
		if err != nil {
			return [][]byte{}, err
		}
		if opts.TerraformConfigFile != "" {
			defer os.Remove(opts.TerraformConfigFile)
		}

		out, err := p.runShow(opts, spinner, "")
		if err != nil {
			return outs, err
		}
		outs = append(outs, out)
	}

	return outs, nil
}

func (p *DirProvider) generatePlanJSONs(paths []string) ([][]byte, error) {
	err := p.checks()
	if err != nil {
		return [][]byte{}, err
	}

	opts, err := p.buildCommandOpts(p.Path)
	if err != nil {
		return [][]byte{}, err
	}
	if opts.TerraformConfigFile != "" {
		defer os.Remove(opts.TerraformConfigFile)
	}

	spinner := ui.NewSpinner("Running terragrunt run-all plan", p.spinnerOpts)
	planFile, planJSON, err := p.runPlan(opts, spinner, true)
	defer os.Remove(planFile)

	if err != nil {
		return [][]byte{}, err
	}

	if len(planJSON) > 0 {
		return [][]byte{planJSON}, nil
	}

	outs := make([][]byte, 0, len(paths))
	spinnerMsg := "Running terragrunt show"
	if len(paths) > 1 {
		spinnerMsg += " for each project"
	}
	spinner = ui.NewSpinner(spinnerMsg, p.spinnerOpts)

	for _, path := range paths {
		opts, err := p.buildCommandOpts(path)
		if err != nil {
			return [][]byte{}, err
		}
		if opts.TerraformConfigFile != "" {
			defer os.Remove(opts.TerraformConfigFile)
		}

		out, err := p.runShow(opts, spinner, planFile)
		if err != nil {
			return outs, err
		}
		outs = append(outs, out)
	}

	return outs, nil
}
