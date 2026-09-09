package terraform

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/infracost/infracost/internal/clierror"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/ui"
)

func findTerraformDir(startDir string) (string, bool) {
	dir := startDir
	for {
		if IsTerraformDir(dir) {
			return dir, true
		}
		parent := filepath.Dir(dir)
		// Stop when reaching root or no further parent
		if parent == dir || parent == "." || parent == "/" {
			break
		}
		dir = parent
	}
	return "", false
}

type PlanProvider struct {
	*DirProvider
	Path                 string
	cachedPlanJSON       []byte
	includePastResources bool
}

func NewPlanProvider(ctx *config.ProjectContext, includePastResources bool) schema.Provider {
	dirProvider := NewDirProvider(ctx, includePastResources).(*DirProvider)

	return &PlanProvider{
		DirProvider:          dirProvider,
		Path:                 ctx.ProjectConfig.Path,
		includePastResources: includePastResources,
	}
}

func (p *PlanProvider) ProjectName() string {
	return config.CleanProjectName(p.ctx.ProjectConfig.Path)
}

func (p *PlanProvider) VarFiles() []string {
	return nil
}

func (p *PlanProvider) RelativePath() string {
	return p.ctx.ProjectConfig.Path
}

func (p *PlanProvider) Type() string {
	return "terraform_plan_binary"
}

func (p *PlanProvider) DisplayType() string {
	return "Terraform plan binary file"
}

func (p *PlanProvider) LoadResources(usage schema.UsageMap) (projects []*schema.Project, err error) {
	j, err := p.generatePlanJSON()
	if err != nil {
		return []*schema.Project{}, err
	}

	logging.Logger.Debug().Msg("Extracting only cost-related params from terraform")
	defer func() {
		if err != nil {
			logging.Logger.Debug().Err(err).Msg("Error running plan provider")
		} else {
			logging.Logger.Debug().Msg("Finished running plan provider")
		}
	}()

	metadata := schema.DetectProjectMetadata(p.ctx.ProjectConfig.Path)
	metadata.Type = p.Type()
	p.AddMetadata(metadata)
	name := p.ctx.ProjectConfig.Name
	if name == "" {
		name = metadata.GenerateProjectName(p.ctx.RunContext.VCSMetadata.Remote, p.ctx.RunContext.IsCloudEnabled())
	}

	project := schema.NewProject(name, metadata)
	parser := NewParser(p.ctx, p.includePastResources)

	j, _ = StripSetupTerraformWrapper(j)
	parsedConf, err := parser.parseJSON(j, usage)
	if err != nil {
		return []*schema.Project{project}, errors.Wrap(err, "Error parsing Terraform JSON")
	}

	project.AddProviderMetadata(parsedConf.ProviderMetadata)

	project.PartialPastResources = parsedConf.PastResources
	project.PartialResources = parsedConf.CurrentResources

	return []*schema.Project{project}, nil
}

func (p *PlanProvider) generatePlanJSON() ([]byte, error) {
	if p.cachedPlanJSON != nil {
		return p.cachedPlanJSON, nil
	}
	if p.cachedPlanJSON != nil {
		logging.Logger.Debug().Msg("Using cached plan JSON")
		return p.cachedPlanJSON, nil
	}

	dir := filepath.Dir(p.Path)
	planPath := filepath.Base(p.Path)

	if !IsTerraformDir(dir) {
		logging.Logger.Debug().Msgf("%s is not a Terraform directory; attempting to find a Terraform directory in parent paths", dir)

		// Try parent directories of the plan file path.
		if parentDir, ok := findTerraformDir(dir); ok {
			logging.Logger.Debug().Msgf("Found Terraform directory in parent path: %s", parentDir)
			dir = parentDir
		} else {
			// Fallback to current working directory (original behaviour)
			cwd, err := os.Getwd()
			if err != nil {
				return []byte{}, err
			}
			logging.Logger.Debug().Msgf("Checking current working directory: %s", cwd)
			planPath = p.Path
			if !IsTerraformDir(cwd) {
				m := fmt.Sprintf("%s %s.\n%s\n\n%s\n%s\n%s %s\n\n%s",
					"Could not detect Terraform directory for",
					p.Path,
					"Either the current working directory, the plan file's parent directory, or one of its parent directories must be a Terraform directory.",
					"If the above does not work you can generate the plan JSON file with:",
					ui.PrimaryString("terraform show -json tfplan.binary > plan.json"),
					"and then run Infracost with",
					ui.PrimaryString("--path=plan.json"),
					"Hint: Ensure the path points to a directory containing Terraform config files (e.g., *.tf) or a plan JSON file.",
				)
				return []byte{}, clierror.NewCLIError(errors.New(m), "Could not detect Terraform directory for plan file")
			}
			dir = cwd
		}
	}

	err := p.checks()
	if err != nil {
		return []byte{}, err
	}

	opts, err := p.buildCommandOpts(dir)
	if err != nil {
		return []byte{}, err
	}
	if opts.TerraformConfigFile != "" {
		defer os.Remove(opts.TerraformConfigFile)
	}

	logging.Logger.Debug().Msg("Running terraform show")

	j, err := p.runShow(opts, planPath, false)
	if err == nil {
		p.cachedPlanJSON = j
	}
	return j, err
}
