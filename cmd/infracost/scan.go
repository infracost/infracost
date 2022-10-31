package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"text/tabwriter"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/hcl"
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

type resourceData struct {
	result   gjson.Result
	resource *schema.Resource
	setKey   string
}

func (s ScanCommand) run(runCtx *config.RunContext) error {
	err := s.loadRunFlags(runCtx.Config)
	if err != nil {
		return err
	}

	var jsons []projectJSON
	for _, project := range runCtx.Config.Projects {
		projectCtx := config.NewProjectContext(runCtx, project, log.Fields{})
		hclProvider, err := terraform.NewHCLProvider(projectCtx, nil, hcl.OptionWithSpinner(projectCtx.RunContext.NewSpinner))
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

	client := http.Client{}
	for _, j := range jsons {
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

		masterProject, err := j.JSONProvider.LoadResourcesFromSrc(map[string]*schema.UsageData{}, j.HCL.JSON, nil)
		if err != nil {
			return err
		}
		schema.BuildResources([]*schema.Project{masterProject}, nil)
		if err := prices.PopulatePrices(runCtx, masterProject); err != nil {
			return err
		}
		schema.CalculateCosts(masterProject)

		parsed := gjson.ParseBytes(j.HCL.JSON)
		resourceMap := make(map[string]resourceData)
		getResourceData("planned_values.root_module", resourceMap, parsed.Get("planned_values.root_module"))

		for _, resource := range masterProject.Resources {
			v := resourceMap[resource.Name]
			resourceMap[resource.Name] = resourceData{
				result:   v.result,
				resource: resource,
				setKey:   v.setKey,
			}
		}

		var maxLen int
		var lines []string
		for _, res := range result.Result {
			cost := "?"

			if !res.NoCost {
				resourceResult := resourceMap[res.Address].result.String()
				value := resourceMap[res.Address].resource
				setKey := resourceMap[res.Address].setKey

				childJSON := parsed.String()
				resourceResult, err := sjson.Set(resourceResult, "values."+res.Attribute, res.Suggested)
				if err != nil {
					return err
				}
				childJSON, _ = sjson.SetRaw(childJSON, setKey, resourceResult)

				projectWithSuggestion, err := j.JSONProvider.LoadResourcesFromSrc(map[string]*schema.UsageData{}, []byte(childJSON), nil)
				if err != nil {
					return err
				}
				schema.BuildResources([]*schema.Project{projectWithSuggestion}, nil)
				if err := prices.PopulatePrices(runCtx, projectWithSuggestion); err != nil {
					return err
				}
				schema.CalculateCosts(projectWithSuggestion)

				var suggested *schema.Resource
				for _, r := range projectWithSuggestion.Resources {
					if r.Name == value.Name {
						suggested = r
						break
					}
				}

				if suggested == nil {
					continue
				}

				diff := decimal.Zero
				if value.MonthlyCost != nil {
					diff = value.MonthlyCost.Sub(*suggested.MonthlyCost)
				}

				cost = output.Format2DP(runCtx.Config.Currency, &diff)
			}

			suggestion := fmt.Sprintf("%q -> %q", res.Current, res.Suggested)
			if strings.Contains(res.Suggested, "+") {
				suggestion = res.Suggested
			}

			line := fmt.Sprintf(
				"%s\t%s\t%s\t%s\t%s",
				res.Address,
				res.Attribute,
				res.Reason,
				suggestion,
				cost,
			)
			if len(line) > maxLen {
				maxLen = len(line)
			}
			lines = append(lines, line)
		}

		fmt.Fprintln(s.cmd.ErrOrStderr())
		fmt.Fprintln(s.cmd.ErrOrStderr(), j.HCL.Module.ModulePath)
		fmt.Fprintln(s.cmd.ErrOrStderr(), strings.Repeat("-", maxLen+12))
		fmt.Fprintln(s.cmd.ErrOrStderr())

		w := tabwriter.NewWriter(s.cmd.ErrOrStderr(), 0, 0, 5, ' ', tabwriter.TabIndent)
		fmt.Fprintln(w, "address\tattribute\treason\tsuggestion\tcost saving")
		fmt.Fprintln(w, "-------\t---------\t------\t----------\t-----------")
		for _, line := range lines {
			fmt.Fprintln(w, line)
		}
		w.Flush()
	}
	return nil
}

func getResourceData(parentKey string, resourceMap map[string]resourceData, module gjson.Result) {
	resources := module.Get("resources")
	for i, resource := range resources.Array() {
		resourceMap[resource.Get("address").String()] = resourceData{
			result: resource,
			setKey: fmt.Sprintf("%s.resources.%d", parentKey, i),
		}
	}

	modules := module.Get("child_modules").Array()
	for i, module := range modules {
		getResourceData(fmt.Sprintf("%s.child_modules.%d", parentKey, i), resourceMap, module)
	}
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
	Address   string `json:"address"`
	Attribute string `json:"attribute"`
	Current   string `json:"current"`
	Suggested string `json:"suggested"`
	Reason    string `json:"reason"`
	NoCost    bool   `json:"no_cost"`
}
