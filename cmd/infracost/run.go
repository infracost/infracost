package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/output"
	"github.com/infracost/infracost/internal/prices"
	"github.com/infracost/infracost/internal/providers/terraform"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/ui"
	"github.com/infracost/infracost/internal/usage"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func addDeprecatedRunFlags(cmd *cobra.Command) {
	cmd.Flags().String("tfjson", "", "Path to Terraform plan JSON file")
	_ = cmd.Flags().MarkHidden("tfjson")

	cmd.Flags().String("tfplan", "", "Path to Terraform plan file relative to 'terraform-dir'")
	_ = cmd.Flags().MarkHidden("tfplan")

	cmd.Flags().String("tfflags", "", "Flags to pass to the 'terraform plan' command")
	_ = cmd.Flags().MarkHidden("tfflags")

	cmd.Flags().String("tfdir", "", "Path to the Terraform code directory. Defaults to current working directory")
	_ = cmd.Flags().MarkHidden("tfdir")

	cmd.Flags().Bool("use-tfstate", false, "Use Terraform state instead of generating a plan")
	_ = cmd.Flags().MarkHidden("use-tfstate")

	cmd.Flags().StringP("output", "o", "table", "Output format: json, table, html")
	_ = cmd.Flags().MarkHidden("output")

	cmd.Flags().String("pricing-api-endpoint", "", "Specify an alternate Cloud Pricing API URL")
	_ = cmd.Flags().MarkHidden("pricing-api-endpoint")
}

func addRunInputFlags(cmd *cobra.Command) {
	cmd.Flags().String("config-file", "", "Path to the Infracost config file. Cannot be used with other flags")
	cmd.Flags().String("usage-file", "", "Path to Infracost usage file that specifies values for usage-based resources")
	cmd.Flags().String("terraform-json-file", "", "Path to Terraform plan JSON file")
	cmd.Flags().String("terraform-plan-file", "", "Path to Terraform plan file relative to 'terraform-dir'")
	cmd.Flags().String("terraform-plan-flags", "", "Flags to pass to the 'terraform plan' command")
	cmd.Flags().String("terraform-dir", "", "Path to the Terraform code directory. Defaults to current working directory")
}

func addRunOutputFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("show-skipped", false, "Show unsupported resources, some of which might be free. Ignored for JSON outputs")
}

func runMain(cfg *config.Config) error {
	projects := make([]*schema.Project, 0)

	for _, projectCfg := range cfg.Projects.Terraform {
		m := fmt.Sprintf("Loading resources from %s", projectCfg.DisplayName())
		if projectCfg.Workspace != "" {
			m += fmt.Sprintf(" (%s)", projectCfg.Workspace)
		}
		if cfg.IsLogging() {
			log.Info(m)
		} else {
			fmt.Fprintln(os.Stderr, m)
		}

		cfg.Environment.SetTerraformEnvironment(projectCfg)

		provider := terraform.New(cfg, projectCfg)

		u, err := usage.LoadFromFile(projectCfg.UsageFile)
		if err != nil {
			return err
		}
		if len(u) > 0 {
			cfg.Environment.HasUsageFile = true
		}

		project, err := provider.LoadResources(u)
		if err != nil {
			return err
		}

		projects = append(projects, project)
	}

	spinnerOpts := ui.SpinnerOptions{
		EnableLogging: cfg.IsLogging(),
		NoColor:       cfg.NoColor,
	}
	spinner := ui.NewSpinner("Calculating cost estimate", spinnerOpts)

	for _, project := range projects {
		if err := prices.PopulatePrices(cfg, project); err != nil {
			spinner.Fail()
			fmt.Fprintln(os.Stderr, "")

			if e := unwrapped(err); errors.Is(e, prices.ErrInvalidAPIKey) {
				return errors.New(fmt.Sprintf("%v\n%s %s %s %s %s\n%s",
					e.Error(),
					"Please check your",
					ui.PrimaryString(config.CredentialsFilePath()),
					"file or",
					ui.PrimaryString("INFRACOST_API_KEY"),
					"environment variable.",
					"If you continue having issues please email hello@infracost.io",
				))
			}

			if e, ok := err.(*prices.PricingAPIError); ok {
				return errors.New(fmt.Sprintf("%v\n%s", e.Error(), "We have been notified of this issue."))
			}

			return err
		}

		schema.CalculateCosts(project)
		project.CalculateDiff()
	}

	spinner.Success()

	r := output.ToOutputFormat(projects)

	for _, outputCfg := range cfg.Outputs {
		cfg.Environment.SetOutputEnvironment(outputCfg)

		opts := output.Options{
			ShowSkipped: outputCfg.ShowSkipped,
			NoColor:     cfg.NoColor,
		}

		var (
			b   []byte
			out string
			err error
		)

		switch strings.ToLower(outputCfg.Format) {
		case "json":
			b, err = output.ToJSON(r, opts)
			out = string(b)
		case "html":
			b, err = output.ToHTML(r, opts)
			out = string(b)
		case "diff":
			b, err = output.ToDiff(r, opts)
			out = fmt.Sprintf("\n%s", string(b))
		case "table_deprecated":
			b, err = output.ToTableDeprecated(r, opts)
			out = fmt.Sprintf("\n%s", string(b))
		default:
			b, err = output.ToTable(r, opts)
			out = fmt.Sprintf("\n%s", string(b))
		}

		if err != nil {
			return errors.Wrap(err, "Error generating output")
		}

		if outputCfg.Path != "" {
			err := ioutil.WriteFile(outputCfg.Path, []byte(out), 0644) // nolint:gosec
			if err != nil {
				return errors.Wrap(err, "Error saving output")
			}
		} else {
			fmt.Printf("%s\n", out)
		}
	}

	return nil
}

func loadRunFlags(cfg *config.Config, cmd *cobra.Command) error {
	cmd.Flags().Changed("terraform-dir")

	hasProjectFlags := (cmd.Flags().Changed("terraform-dir") ||
		cmd.Flags().Changed("terraform-plan-file") ||
		cmd.Flags().Changed("terraform-json-file") ||
		cmd.Flags().Changed("terraform-use-state") ||
		cmd.Flags().Changed("terraform-plan-flags") ||
		cmd.Flags().Changed("usage-file"))

	hasOutputFlags := (cmd.Flags().Changed("format") ||
		cmd.Flags().Changed("show-skipped"))

	if cmd.Flags().Changed("config-file") {
		if hasProjectFlags || hasOutputFlags {
			ui.PrintUsageErrorAndExit(cmd, "--config-file flag cannot be used with other project and output flags")
		}

		configFile, _ := cmd.Flags().GetString("config-file")
		return cfg.LoadFromFile(configFile)
	}

	projectCfg := &config.TerraformProject{}
	outputCfg := &config.Output{}

	if hasProjectFlags {
		cfg.Projects = config.Projects{
			Terraform: []*config.TerraformProject{
				projectCfg,
			},
		}
	}

	if hasOutputFlags {
		cfg.Outputs = []*config.Output{outputCfg}
	}

	if hasProjectFlags || hasOutputFlags {
		err := cfg.LoadFromEnv()
		if err != nil {
			return err
		}
	}

	if hasProjectFlags {
		projectCfg.Dir, _ = cmd.Flags().GetString("terraform-dir")
		projectCfg.PlanFile, _ = cmd.Flags().GetString("terraform-plan-file")
		projectCfg.JSONFile, _ = cmd.Flags().GetString("terraform-json-file")
		projectCfg.UseState, _ = cmd.Flags().GetBool("terraform-use-state")
		projectCfg.PlanFlags, _ = cmd.Flags().GetString("terraform-plan-flags")
		projectCfg.UsageFile, _ = cmd.Flags().GetString("usage-file")
	}

	if hasOutputFlags {
		outputCfg.Format, _ = cmd.Flags().GetString("format")
		outputCfg.ShowSkipped, _ = cmd.Flags().GetBool("show-skipped")
	}

	return nil
}

func checkRunConfig(cfg *config.Config) error {
	for _, projectCfg := range cfg.Projects.Terraform {
		if projectCfg.UseState && (projectCfg.PlanFile != "" || projectCfg.JSONFile != "") {
			return errors.New("The use state option cannot be used with the Terraform plan or Terraform JSON options")
		}

		if projectCfg.JSONFile != "" && projectCfg.PlanFile != "" {
			return errors.New("Please provide either a Terraform Plan JSON file or a Terraform Plan file")
		}

		if projectCfg.Dir != "" && projectCfg.JSONFile != "" {
			ui.PrintWarning("Warning: Terraform directory is ignored if Terraform JSON is used")
			return nil
		}
	}

	for _, output := range cfg.Outputs {
		if output.Format == "json" && output.ShowSkipped {
			ui.PrintWarning("The show skipped option is not needed with JSON output as that always includes them.")
			return nil
		}
	}

	return nil
}

func unwrapped(err error) error {
	e := err
	for errors.Unwrap(e) != nil {
		e = errors.Unwrap(e)
	}

	return e
}
