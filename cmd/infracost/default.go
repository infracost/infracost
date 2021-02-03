package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/fatih/color"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/output"
	"github.com/infracost/infracost/internal/prices"
	"github.com/infracost/infracost/internal/providers/terraform"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/spin"
	"github.com/infracost/infracost/internal/usage"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func defaultCmd() *cli.Command {
	deprecatedFlags := []cli.Flag{
		&cli.StringFlag{
			Name:      "tfjson",
			Usage:     "Path to Terraform plan JSON file",
			TakesFile: true,
			Hidden:    true,
		},
		&cli.StringFlag{
			Name:      "tfplan",
			Usage:     "Path to Terraform plan file relative to 'tfdir'",
			TakesFile: true,
			Hidden:    true,
		},
		&cli.BoolFlag{
			Name:   "use-tfstate",
			Usage:  "Use Terraform state instead of generating a plan",
			Value:  false,
			Hidden: true,
		},
		&cli.StringFlag{
			Name:        "tfdir",
			Usage:       "Path to the Terraform code directory",
			TakesFile:   true,
			DefaultText: "current working directory",
			Hidden:      true,
		},
		&cli.StringFlag{
			Name:   "tfflags",
			Usage:  "Flags to pass to the 'terraform plan' command",
			Hidden: true,
		},
		&cli.StringFlag{
			Name:   "output",
			Usage:  "Output format: json, table, html",
			Value:  "table",
			Hidden: true,
		},
	}

	deprecatedFlagsMapping := map[string]string{
		"tfjson":      "terraform-json-file",
		"tfplan":      "terraform-plan-file",
		"use-tfstate": "terraform-use-state",
		"tfdir":       "terraform-dir",
		"tfflags":     "terraform-plan-flags",
		"output":      "format",
		"o":           "format",
	}

	flags := append(deprecatedFlags,
		&cli.StringFlag{
			Name:      "terraform-json-file",
			Usage:     "Path to Terraform plan JSON file",
			TakesFile: true,
		},
		&cli.StringFlag{
			Name:      "terraform-plan-file",
			Usage:     "Path to Terraform plan file relative to 'terraform-dir'",
			TakesFile: true,
		},
		&cli.BoolFlag{
			Name:  "terraform-use-state",
			Usage: "Use Terraform state instead of generating a plan",
			Value: false,
		},
		&cli.StringFlag{
			Name:        "terraform-dir",
			Usage:       "Path to the Terraform code directory",
			TakesFile:   true,
			DefaultText: "current working directory",
		},
		&cli.StringFlag{
			Name:  "terraform-plan-flags",
			Usage: "Flags to pass to the 'terraform plan' command",
		},
		&cli.StringFlag{
			Name:   "format",
			Usage:  "Output format: json, table, html",
			Value:  "table",
			Hidden: true,
		},
		&cli.BoolFlag{
			Name:  "show-skipped",
			Usage: "Show unsupported resources, some of which might be free. Only for table and HTML output format",
			Value: false,
		},
		&cli.StringFlag{
			Name:      "usage-file",
			Usage:     "Path to Infracost usage file that specifies values for usage-based resources",
			TakesFile: true,
		},
	)

	return &cli.Command{
		Flags: flags,
		Action: func(c *cli.Context) error {
			if err := checkAPIKey(); err != nil {
				return err
			}

			handleDeprecatedFlags(c, deprecatedFlagsMapping)
			loadDefaultCmdFlags(c)
			err := checkUsageErrors()
			if err != nil {
				usageError(c, err.Error())
			}

			return defaultMain()
		},
	}
}

func handleDeprecatedFlags(c *cli.Context, deprecatedFlagsMapping map[string]string) {
	for _, flagName := range c.FlagNames() {
		if newName, ok := deprecatedFlagsMapping[flagName]; ok {
			m := fmt.Sprintf("Flag --%s is deprecated and will be removed in v0.8.0.", flagName)
			if newName != "" {
				m += fmt.Sprintf(" Please use --%s.", newName)
			}

			usageWarning(m)

			if !c.IsSet(newName) {
				err := c.Set(newName, c.String(flagName))
				if err != nil {
					log.Debugf("Error setting flag %s from %s", newName, flagName)
				}
			}
		}
	}
}

func loadDefaultCmdFlags(c *cli.Context) {
	useProjectFlags := false
	useOutputFlags := false

	project := &config.TerraformProjectSpec{}

	if c.IsSet("terraform-dir") {
		useProjectFlags = true
		project.Dir = c.String("terraform-dir")
	}

	if c.IsSet("terraform-plan-file") {
		useProjectFlags = true
		project.PlanFile = c.String("terraform-plan-file")
	}

	if c.IsSet("terraform-json-file") {
		useProjectFlags = true
		project.JSONFile = c.String("terraform-json-file")
	}

	if c.IsSet("terraform-use-state") {
		useProjectFlags = true
		project.UseState = c.Bool("terraform-use-state")
	}

	if c.IsSet("terraform-plan-flags") {
		useProjectFlags = true
		project.PlanFlags = c.String("terraform-plan-flags")
	}

	if c.IsSet("usage-file") {
		useProjectFlags = true
		project.UsageFile = c.String("usage-file")
	}

	if useProjectFlags {
		config.Config.Projects = config.ProjectSpec{
			Terraform: []*config.TerraformProjectSpec{project},
		}
	}

	output := &config.OutputSpec{}

	if c.IsSet("format") {
		useOutputFlags = true
		output.Format = c.String("format")
	}

	if c.IsSet("show-skipped") {
		useOutputFlags = true
		output.ShowSkipped = c.Bool("show-skipped")
	}

	if useOutputFlags {
		config.Config.Outputs = []*config.OutputSpec{output}
	}
}

func checkUsageErrors() error {
	for _, project := range config.Config.Projects.Terraform {
		if project.UseState && (project.PlanFile != "" || project.JSONFile != "") {
			return errors.New("The use state option cannot be used with the Terraform plan or Terraform JSON options")
		}

		if project.JSONFile != "" && project.PlanFile != "" {
			return errors.New("Please provide either a Terraform Plan JSON file or a Terraform Plan file")
		}

		if project.Dir != "" && project.JSONFile != "" {
			usageWarning("Warning: Terraform directory is ignored if Terraform JSON is used")
			return nil
		}
	}

	for _, output := range config.Config.Outputs {
		if output.Format == "json" && output.ShowSkipped {
			usageWarning("The show skipped option is not needed with JSON output as that always includes them.")
			return nil
		}
	}

	return nil
}

func defaultMain() error {

	resources := make([]*schema.Resource, 0)

	for _, project := range config.Config.Projects.Terraform {
		provider := terraform.New(project)

		u, err := usage.LoadFromFile(project.UsageFile)
		if err != nil {
			return err
		}
		if len(u) > 0 {
			config.Environment.HasUsageFile = true
		}

		r, err := provider.LoadResources(u)
		if err != nil {
			return err
		}
		resources = append(resources, r...)
	}

	spinner = spin.NewSpinner("Calculating cost estimate")

	if err := prices.PopulatePrices(resources); err != nil {
		spinner.Fail()

		red := color.New(color.FgHiRed)
		bold := color.New(color.Bold, color.FgHiWhite)

		if e := unwrapped(err); errors.Is(e, prices.ErrInvalidAPIKey) {
			return errors.New(fmt.Sprintf("%v\n%s %s %s %s %s\n%s",
				e.Error(),
				red.Sprint("Please check your"),
				bold.Sprint(config.CredentialsFilePath()),
				red.Sprint("file or"),
				bold.Sprint("INFRACOST_API_KEY"),
				red.Sprint("environment variable."),
				red.Sprint("If you continue having issues please email hello@infracost.io"),
			))
		}

		if e, ok := err.(*prices.PricingAPIError); ok {
			return errors.New(fmt.Sprintf("%v\n%s", e.Error(), "We have been notified of this issue."))
		}

		return err
	}

	schema.CalculateCosts(resources)

	spinner.Success()

	schema.SortResources(resources)

	opts := output.Options{}
	r := output.ToOutputFormat(resources)

	for _, outputConfig := range config.Config.Outputs {
		var (
			b   []byte
			out string
			err error
		)

		switch strings.ToLower(outputConfig.Format) {
		case "json":
			b, err = output.ToJSON(r)
			out = string(b)
		case "html":
			b, err = output.ToHTML(r, opts, outputConfig)
			out = string(b)
		default:
			b, err = output.ToTable(r, outputConfig)
			out = fmt.Sprintf("\n%s", string(b))
		}

		if err != nil {
			return errors.Wrap(err, "Error generating output")
		}

		if outputConfig.Path != "" {
			err := ioutil.WriteFile(outputConfig.Path, []byte(out), 0644) // nolint:gosec
			if err != nil {
				return errors.Wrap(err, "Error saving output")
			}
		} else {
			fmt.Printf("%s\n", out)
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
