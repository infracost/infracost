package main

import (
	"fmt"
	"io/ioutil"
	"os"
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

var deprecatedFlagsMapping = map[string]string{
	"tfjson":      "terraform-json-file",
	"tfplan":      "terraform-plan-file",
	"use-tfstate": "terraform-use-state",
	"tfdir":       "terraform-dir",
	"tfflags":     "terraform-plan-flags",
	"output":      "format",
	"o":           "format",
}

var deprecatedEnvVarMapping = map[string]string{
	"TERRAFORM_BINARY":      "INFRACOST_TERRAFORM_BINARY",
	"TERRAFORM_CLOUD_HOST":  "INFRACOST_TERRAFORM_CLOUD_HOST",
	"TERRAFORM_CLOUD_TOKEN": "INFRACOST_TERRAFORM_CLOUD_TOKEN",
	"SKIP_UPDATE_CHECK": "INFRACOST_SKIP_UPDATE_CHECK",
}

func defaultCmd(cfg *config.Config) *cli.Command {
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

	flags := append(deprecatedFlags,
		&cli.StringFlag{
			Name:  "config-file",
			Usage: "Path to the Infracost config file. Cannot be used with other flags",
		},
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
			handleDeprecatedEnvVars(c, deprecatedEnvVarMapping)
			handleDeprecatedFlags(c, deprecatedFlagsMapping)

			if err := checkAPIKey(cfg.APIKey, cfg.PricingAPIEndpoint, cfg.DefaultPricingAPIEndpoint); err != nil {
				return err
			}

			err := loadDefaultCmdFlags(cfg, c)
			if err != nil {
				return err
			}

			err = checkUsageErrors(cfg)
			if err != nil {
				usageError(c, err.Error())
			}

			return defaultMain(cfg)
		},
	}
}

func handleDeprecatedEnvVars(c *cli.Context, deprecatedEnvVars map[string]string) {
	for oldName, newName := range deprecatedEnvVars {
		if val, ok := os.LookupEnv(oldName); ok {
			m := fmt.Sprintf("Environment variable %s is deprecated and will be removed in v0.8.0.", oldName)
			if newName != "" {
				m += fmt.Sprintf(" Please use %s.", newName)
			}

			usageWarning(m)

			if _, ok := os.LookupEnv(newName); !ok {
				os.Setenv(newName, val)
			}
		}
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

func loadDefaultCmdFlags(cfg *config.Config, c *cli.Context) error {
	hasProjectFlags := (c.IsSet("terraform-dir") ||
		c.IsSet("terraform-plan-file") ||
		c.IsSet("terraform-json-file") ||
		c.IsSet("terraform-use-state") ||
		c.IsSet("terraform-plan-flags") ||
		c.IsSet("usage-file"))

	hasOutputFlags := (c.IsSet("format") ||
		c.IsSet("show-skipped"))

	if c.IsSet("config-file") {
		if hasProjectFlags || hasOutputFlags {
			usageError(c, "--config-file flag cannot be used with other flags")
		}

		return cfg.LoadFromFile(c.String("config-file"))
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
		projectCfg.Dir = c.String("terraform-dir")
		projectCfg.PlanFile = c.String("terraform-plan-file")
		projectCfg.JSONFile = c.String("terraform-json-file")
		projectCfg.UseState = c.Bool("terraform-use-state")
		projectCfg.PlanFlags = c.String("terraform-plan-flags")
		projectCfg.UsageFile = c.String("usage-file")
	}

	if hasOutputFlags {
		outputCfg.Format = c.String("format")
		outputCfg.ShowSkipped = c.Bool("show-skipped")
	}

	return nil
}

func checkUsageErrors(cfg *config.Config) error {
	for _, projectCfg := range cfg.Projects.Terraform {
		if projectCfg.UseState && (projectCfg.PlanFile != "" || projectCfg.JSONFile != "") {
			return errors.New("The use state option cannot be used with the Terraform plan or Terraform JSON options")
		}

		if projectCfg.JSONFile != "" && projectCfg.PlanFile != "" {
			return errors.New("Please provide either a Terraform Plan JSON file or a Terraform Plan file")
		}

		if projectCfg.Dir != "" && projectCfg.JSONFile != "" {
			usageWarning("Warning: Terraform directory is ignored if Terraform JSON is used")
			return nil
		}
	}

	for _, output := range cfg.Outputs {
		if output.Format == "json" && output.ShowSkipped {
			usageWarning("The show skipped option is not needed with JSON output as that always includes them.")
			return nil
		}
	}

	return nil
}

func defaultMain(cfg *config.Config) error {
	resources := make([]*schema.Resource, 0)

	for _, projectCfg := range cfg.Projects.Terraform {
		dir := projectCfg.Dir
		if dir == "." || dir == "" {
			dir = "current directory"
		}

		m := fmt.Sprintf("Loading resources from %s", dir)
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

		r, err := provider.LoadResources(u)
		if err != nil {
			return err
		}
		resources = append(resources, r...)
	}

	spinnerOpts := spin.Options{
		EnableLogging: cfg.IsLogging(),
		NoColor:       cfg.NoColor,
	}
	spinner := spin.NewSpinner("Calculating cost estimate", spinnerOpts)

	if err := prices.PopulatePrices(cfg, resources); err != nil {
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

	r := output.ToOutputFormat(resources)

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

func unwrapped(err error) error {
	e := err
	for errors.Unwrap(e) != nil {
		e = errors.Unwrap(e)
	}

	return e
}
