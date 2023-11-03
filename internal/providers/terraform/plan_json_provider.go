package terraform

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"

	"github.com/infracost/infracost/internal/apiclient"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/ui"
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
	spinner := ui.NewSpinner("Extracting only cost-related params from terraform", ui.SpinnerOptions{
		EnableLogging: p.ctx.RunContext.Config.IsLogging(),
		NoColor:       p.ctx.RunContext.Config.NoColor,
		Indent:        "  ",
	})
	defer spinner.Fail()

	j, err := os.ReadFile(p.Path)
	if err != nil {
		return []*schema.Project{}, fmt.Errorf("Error reading Terraform plan JSON file %w", err)
	}

	project, err := p.LoadResourcesFromSrc(usage, j, spinner)
	if err != nil {
		return nil, err
	}

	return []*schema.Project{project}, nil
}

func (p *PlanJSONProvider) LoadResourcesFromSrc(usage schema.UsageMap, j []byte, spinner *ui.Spinner) (*schema.Project, error) {
	metadata := config.DetectProjectMetadata(p.ctx.ProjectConfig.Path)
	metadata.Type = p.Type()
	p.AddMetadata(metadata)
	name := p.ctx.ProjectConfig.Name
	if name == "" {
		name = metadata.GenerateProjectName(p.ctx.RunContext.VCSMetadata.Remote, p.ctx.RunContext.IsCloudEnabled())
	}

	project := schema.NewProject(name, metadata)
	parser := NewParser(p.ctx, p.includePastResources)

	partialPastResources, partialResources, providerMetadatas, err := parser.parseJSON(j, usage)
	if err != nil {
		return project, fmt.Errorf("Error parsing Terraform plan JSON file %w", err)
	}

	project.AddProviderMetadata(providerMetadatas)

	project.PartialPastResources = partialPastResources
	project.PartialResources = partialResources

	// use TagPolicyAPIEndpoint for Policy2 instead of creating a new config variable
	if p.policyClient != nil {
		err := p.policyClient.UploadPolicyData(project)
		if err != nil {
			p.logger.Err(err).Msgf("Terraform project %s failed to upload policy data", project.Name)
		}
	}

	if spinner != nil {
		spinner.Success()
	}

	return project, nil
}
