package terraform

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"

	"github.com/infracost/infracost/internal/apiclient"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/schema"
)

type PlanJSONProvider struct {
	ctx                  *config.ProjectContext
	Path                 string
	includePastResources bool
	policyClient         *apiclient.PolicyAPIClient
	logger               zerolog.Logger
}

func NewPlanJSONProvider(ctx *config.ProjectContext, includePastResources bool) *PlanJSONProvider {
	var policyClient *apiclient.PolicyAPIClient
	var err error
	if ctx.RunContext.Config.PoliciesEnabled {
		policyClient, err = apiclient.NewPolicyAPIClient(ctx.RunContext)
		if err != nil {
			logging.Logger.Debug().Err(err).Msgf("failed to initialize policy client")
		}
	}

	return &PlanJSONProvider{
		ctx:                  ctx,
		Path:                 ctx.ProjectConfig.Path,
		includePastResources: includePastResources,
		policyClient:         policyClient,
		logger:               ctx.Logger(),
	}
}

func (p *PlanJSONProvider) ProjectName() string {
	return config.CleanProjectName(p.ctx.ProjectConfig.Path)
}

func (p *PlanJSONProvider) VarFiles() []string {
	return nil
}

func (p *PlanJSONProvider) RelativePath() string {
	return p.ctx.ProjectConfig.Path
}

func (p *PlanJSONProvider) Context() *config.ProjectContext {
	return p.ctx
}

func (p *PlanJSONProvider) Type() string {
	return "terraform_plan_json"
}

func (p *PlanJSONProvider) DisplayType() string {
	return "Terraform plan JSON file"
}

func (p *PlanJSONProvider) AddMetadata(metadata *schema.ProjectMetadata) {
	metadata.ConfigSha = p.ctx.ProjectConfig.ConfigSha

	// TerraformWorkspace isn't used to load resources but we still pass it
	// on so it appears in the project name of the output
	metadata.TerraformWorkspace = p.ctx.ProjectConfig.TerraformWorkspace
}

func (p *PlanJSONProvider) LoadResources(usage schema.UsageMap) ([]*schema.Project, error) {
	j, err := os.ReadFile(p.Path)
	if err != nil {
		return []*schema.Project{}, fmt.Errorf("Error reading Terraform plan JSON file %w", err)
	}

	project, err := p.LoadResourcesFromSrc(usage, j)
	if err != nil {
		return nil, err
	}

	return []*schema.Project{project}, nil
}

func (p *PlanJSONProvider) LoadResourcesFromSrc(usage schema.UsageMap, j []byte) (*schema.Project, error) {
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
		return project, fmt.Errorf("Error parsing Terraform plan JSON file %w", err)
	}

	project.AddProviderMetadata(parsedConf.ProviderMetadata)

	project.PartialPastResources = parsedConf.PastResources
	project.PartialResources = parsedConf.CurrentResources

	// use TagPolicyAPIEndpoint for Policy2 instead of creating a new config variable
	if p.policyClient != nil {
		err := p.policyClient.UploadPolicyData(project, parsedConf.CurrentResourceDatas, parsedConf.PastResourceDatas)
		if err != nil {
			p.logger.Err(err).Msgf("Terraform project %s failed to upload policy data", project.Name)
		}
	}

	return project, nil
}
