package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/imdario/mergo"
	jsoniter "github.com/json-iterator/go"
	"github.com/pterm/pterm"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/infracost/infracost/internal/apiclient"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/output"
	"github.com/infracost/infracost/internal/prices"
	"github.com/infracost/infracost/internal/providers/terraform"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/ui"
)

type ScanCommand struct {
	TerraformVarFiles  []string
	TerraformVars      []string
	TerraformWorkspace string

	Path       string
	ConfigFile string
	UsageFile  string

	cmd *cobra.Command
}

func (s ScanCommand) printUsage() {
	ui.PrintUsage(s.cmd)
}
func (s ScanCommand) hasProjectFlags() bool {
	return s.Path != "" || len(s.TerraformVars) > 0 || len(s.TerraformVarFiles) > 0 || s.TerraformWorkspace != ""
}

func (s ScanCommand) loadRunFlags(cfg *config.Config) error {
	if s.Path == "" && s.ConfigFile == "" {
		m := fmt.Sprintf("No path specified\n\nUse the %s flag to specify the path to one of the following:\n", ui.PrimaryString("--path"))
		m += fmt.Sprintf(" - Terraform/Terragrunt directory\n - Terraform plan JSON file, see %s for how to generate this.", ui.SecondaryLinkString("https://infracost.io/troubleshoot"))
		m += fmt.Sprintf("\n\nAlternatively, use --config-file to process multiple projects, see %s", ui.SecondaryLinkString("https://infracost.io/config-file"))

		s.printUsage()
		return errors.New(m)
	}

	if s.ConfigFile != "" && s.hasProjectFlags() {
		m := "--config-file flag cannot be used with the following flags: "
		m += "--path, --project-name, --terraform-*, --usage-file"
		s.printUsage()
		return errors.New(m)
	}

	projectCfg := cfg.Projects[0]

	if s.hasProjectFlags() {
		cfg.RootPath = s.Path
		projectCfg.Path = s.Path

		projectCfg.TerraformVarFiles = s.TerraformVarFiles
		projectCfg.TerraformVars = tfVarsToMap(s.TerraformVars)
		projectCfg.UsageFile = s.UsageFile
		projectCfg.TerraformWorkspace = s.TerraformWorkspace
	}

	if s.ConfigFile != "" {
		err := cfg.LoadFromConfigFile(s.ConfigFile)
		if err != nil {
			return err
		}

		cfg.ConfigFilePath = s.ConfigFile
	}

	return nil
}

type projectJSON struct {
	HCL          terraform.HclProject
	JSONProvider *terraform.PlanJSONProvider
}

func (s ScanCommand) run(runCtx *config.RunContext) error {
	err := s.loadRunFlags(runCtx.Config)
	if err != nil {
		return err
	}

	var jsons []projectJSON
	spinner, _ := pterm.DefaultSpinner.Start("Detecting Terraform Projects...")
	for _, project := range runCtx.Config.Projects {
		projectCtx := config.NewProjectContext(runCtx, project, log.Fields{})
		hclProvider, err := terraform.NewHCLProvider(projectCtx, &terraform.HCLProviderConfig{SuppressLogging: true})
		if err != nil {
			return err
		}

		projectJsons, err := hclProvider.LoadPlanJSONs()
		if err != nil {
			return err
		}

		planJsonProvider := terraform.NewPlanJSONProvider(projectCtx, false)
		for _, j := range projectJsons {
			jsons = append(jsons, projectJSON{HCL: j, JSONProvider: planJsonProvider})
		}
	}

	pricingClient := apiclient.NewPricingAPIClient(runCtx)
	client := http.Client{Timeout: 5 * time.Second}
	spinner.Success()

	for _, j := range jsons {
		spinner, _ = pterm.DefaultSpinner.Start(fmt.Sprintf("Scanning project %s for cost optimizations...", j.HCL.Module.ModulePath))

		buf := bytes.NewBuffer(j.HCL.JSON)
		req, err := http.NewRequest(http.MethodPost, "http://localhost:8081/recommend", buf)
		if err != nil {
			return err
		}

		res, err := client.Do(req)
		if err != nil {
			return err
		}
		var result RecommendDecisionResponse
		json.NewDecoder(res.Body).Decode(&result)

		if len(result.Result) == 0 {
			continue
		}

		var recMap = make(map[string][]Suggestion)
		for _, suggestion := range result.Result {
			if v, ok := recMap[suggestion.ResourceType]; ok {
				recMap[suggestion.ResourceType] = append(v, suggestion)
				continue
			}

			recMap[suggestion.ResourceType] = []Suggestion{suggestion}
		}

		masterProject, err := j.JSONProvider.LoadResourcesFromSrc(map[string]*schema.UsageData{}, j.HCL.JSON, nil)
		if err != nil {
			return err
		}

		rows := pterm.TableData{
			{"Address", "Reason", "Suggestion", "Cost Saving"},
		}

		for _, resource := range masterProject.PartialResources {
			coreResource := resource.CoreResource
			if coreResource != nil {
				if suggestions, ok := recMap[coreResource.CoreType()]; ok {
					// TODO: fetch usage and populate resource
					coreResource.PopulateUsage(nil)
					initialSchema, err := jsoniter.Marshal(coreResource)
					if err != nil {
						return err
					}

					initalResource := coreResource.BuildResource()
					err = prices.GetPrices(runCtx, pricingClient, initalResource)
					if err != nil {
						return err
					}
					initalResource.CalculateCosts()

					for _, suggestion := range suggestions {
						if suggestion.Address != initalResource.Name {
							continue
						}

						if suggestion.NoCost {
							rows = append(rows, []string{
								suggestion.Address,
								suggestion.Reason,
								suggestion.Suggested,
								"?",
							})
							continue
						}

						suggestedAttributes := suggestion.ResourceAttributes
						err = mergeSuggestionWithResource(initialSchema, suggestedAttributes, coreResource)
						if err != nil {
							return err
						}

						schemaResource := coreResource.BuildResource()
						err = prices.GetPrices(runCtx, pricingClient, schemaResource)
						if err != nil {
							return err
						}
						schemaResource.CalculateCosts()

						diff := decimal.Zero
						if schemaResource.MonthlyCost != nil {
							diff = initalResource.MonthlyCost.Sub(*schemaResource.MonthlyCost)
						}

						cost := output.Format2DP(runCtx.Config.Currency, &diff)

						rows = append(rows, []string{
							suggestion.Address,
							suggestion.Reason,
							suggestion.Suggested,
							cost,
						})
					}

				}
			}
		}

		spinner.Success()
		pterm.DefaultBox.WithTitle(j.HCL.Module.ModulePath).Println(pterm.DefaultTable.WithHasHeader().WithData(rows).Srender())
	}

	return nil
}

func mergeSuggestionWithResource(schema []byte, suggestedSchema []byte, resource schema.CoreResource) error {
	var initialAttributes map[string]interface{}
	jsoniter.Unmarshal(schema, &initialAttributes)

	var suggestedAttributes map[string]interface{}
	jsoniter.Unmarshal(suggestedSchema, &suggestedAttributes)

	err := mergo.Merge(&initialAttributes, suggestedAttributes, mergo.WithOverride, mergo.WithSliceDeepCopy)
	if err != nil {
		return err
	}

	nb, err := jsoniter.Marshal(initialAttributes)
	if err != nil {
		return err
	}
	err = json.Unmarshal(nb, &resource)
	if err != nil {
		return err
	}

	return nil
}

func scanCommand(ctx *config.RunContext) *cobra.Command {
	var scan ScanCommand

	cmd := &cobra.Command{
		Use:    "scan",
		Short:  "Scan your Terraform projects for cost optimisations",
		Long:   "Scan your Terraform projects for cost optimisations",
		Hidden: true,
		Example: `
      infracost scan --path /code --terraform-var-file my.tfvars
      `,
		ValidArgs: []string{"--", "-"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkAPIKey(ctx.Config.APIKey, ctx.Config.PricingAPIEndpoint, ctx.Config.DefaultPricingAPIEndpoint); err != nil {
				return err
			}

			return scan.run(ctx)
		},
	}

	scan.cmd = cmd
	cmd.Flags().StringSliceVar(&scan.TerraformVarFiles, "terraform-var-file", nil, "Load variable files, similar to Terraform's -var-file flag. Provided files must be relative to the --path flag")
	cmd.Flags().StringSliceVar(&scan.TerraformVars, "terraform-var", nil, "Set value for an input variable, similar to Terraform's -var flag")
	cmd.Flags().StringVar(&scan.TerraformWorkspace, "terraform-workspace", "", "Terraform workspace to use. Applicable when path is a Terraform directory")

	cmd.Flags().StringVarP(&scan.Path, "path", "p", "", "Path to the Terraform directory or JSON/plan file")
	cmd.Flags().StringVar(&scan.ConfigFile, "config-file", "", "Path to Infracost config file. Cannot be used with path, terraform* or usage-file flags")
	cmd.Flags().StringVar(&scan.UsageFile, "usage-file", "", "Path to Infracost usage file that specifies values for usage-based resources")

	return cmd
}

type RecommendDecisionResponse struct {
	Result []Suggestion `json:"result"`
}

type Suggestion struct {
	ResourceType       string          `json:"resourceType"`
	ResourceAttributes json.RawMessage `json:"resourceAttributes"`
	Address            string          `json:"address"`
	Reason             string          `json:"reason"`
	Suggested          string          `json:"suggested"`
	NoCost             bool            `json:"no_cost"`
}
