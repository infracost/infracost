package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/ui"
)

var supportedConfigureKeys = map[string]struct{}{
	"api_key":                  {},
	"currency":                 {},
	"pricing_api_endpoint":     {},
	"enable_dashboard":         {},
	"enable_cloud":             {},
	"disable_hcl":              {},
	"tls_insecure_skip_verify": {},
	"tls_ca_cert_file":         {},
}

func configureCmd(ctx *config.RunContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configure",
		Short: "Display or change global configuration",
		Long:  supportedConfigureSettingsOutput("Display or change global configuration"),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Show the help
			return cmd.Help()
		},
	}

	cmd.AddCommand(configureSetCmd(ctx), configureGetCmd(ctx))

	return cmd
}

func configureSetCmd(ctx *config.RunContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set global configuration",
		Long:  supportedConfigureSettingsOutput("Set global configuration"),
		Example: `  Set your Infracost API key:

      infracost configure set api_key MY_API_KEY

  Set your Cloud Pricing API endpoint:

      infracost	configure set pricing_api_endpoint https://cloud-pricing-api

  Set your preferred currency code (ISO 4217):

      infracost	configure set currency EUR
`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 2 {
				return errors.New("Too many arguments")
			}

			if len(args) < 2 {
				return errors.New("You must specify a key and a value")
			}

			if !isValidConfigureKey(args[0]) {
				return fmt.Errorf("Invalid key, valid keys are: %s", strings.Join(validConfigureKeys(), ", "))
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]

			saveCredentials := false
			saveConfiguration := false

			switch key {
			case "pricing_api_endpoint":
				ctx.Config.Credentials.PricingAPIEndpoint = value
				saveCredentials = true
			case "api_key":
				ctx.Config.Credentials.APIKey = value
				saveCredentials = true
			case "tls_insecure_skip_verify":
				if value == "" {
					ctx.Config.Configuration.TLSInsecureSkipVerify = nil
				} else {
					b, err := strconv.ParseBool(value)

					if err != nil {
						return errors.New("Invalid value, must be true or false")
					}

					ctx.Config.Configuration.TLSInsecureSkipVerify = &b
				}
				saveConfiguration = true
			case "tls_ca_cert_file":
				ctx.Config.Configuration.TLSCACertFile = value
				saveConfiguration = true
			case "currency":
				ctx.Config.Configuration.Currency = value
				saveConfiguration = true
			case "disable_hcl":
				b, err := strconv.ParseBool(value)
				if err != nil {
					return errors.New("Invalid value, must be true or false")
				}

				ctx.Config.Configuration.DisableHCLParsing = &b
				saveConfiguration = true
			case "enable_dashboard":
				b, err := strconv.ParseBool(value)

				if err != nil {
					return errors.New("Invalid value, must be true or false")
				}

				if b && ctx.Config.IsSelfHosted() {
					return errors.New("The dashboard is part of Infracost's hosted services. Contact hello@infracost.io for help")
				}

				ctx.Config.Configuration.EnableDashboard = &b
				saveConfiguration = true
			case "enable_cloud":
				b, err := strconv.ParseBool(value)

				if err != nil {
					return errors.New("Invalid value, must be true or false")
				}

				if b && ctx.Config.IsSelfHosted() {
					return errors.New("Infracost Cloud is part of Infracost's hosted services. Contact hello@infracost.io for help")
				}

				ctx.Config.Configuration.EnableCloud = &b
				saveConfiguration = true
			}

			if saveCredentials {
				err := ctx.Config.Credentials.Save()
				if err != nil {
					return err
				}
			}

			if saveConfiguration {
				err := ctx.Config.Configuration.Save()
				if err != nil {
					return err
				}
			}

			return nil
		},
	}

	return cmd
}

func configureGetCmd(ctx *config.RunContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Get global configuration",
		Long:  supportedConfigureSettingsOutput("Get global configuration"),
		Example: `  Get your saved Infracost API key:

      infracost configure get api_key

  Get your saved Cloud Pricing API endpoint:

      infracost	configure get pricing_api_endpoint

  Get your preferred currency:

      infracost	configure get currency
`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return errors.New("Too many arguments")
			}

			if len(args) < 1 {
				return errors.New("You must specify a key")
			}

			if !isValidConfigureKey(args[0]) {
				return fmt.Errorf("Invalid key, valid keys are: %s", strings.Join(validConfigureKeys(), ", "))
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			var value string

			switch key {
			case "pricing_api_endpoint":
				value = ctx.Config.Credentials.PricingAPIEndpoint

				if value == "" {
					msg := fmt.Sprintf("No Cloud Pricing API endpoint in your saved config (%s).\nSet an API key using %s.",
						config.CredentialsFilePath(),
						ui.PrimaryString("infracost configure set pricing_api_endpoint https://cloud-pricing-api"),
					)
					logging.Logger.Warn().Msg(msg)
				}
			case "api_key":
				value = ctx.Config.Credentials.APIKey

				if value == "" {
					msg := fmt.Sprintf("No API key in your saved config (%s).\nSet an API key using %s.",
						config.CredentialsFilePath(),
						ui.PrimaryString("infracost configure set api_key MY_API_KEY"),
					)
					logging.Logger.Warn().Msg(msg)
				}
			case "currency":
				value = ctx.Config.Configuration.Currency

				if value == "" {
					msg := fmt.Sprintf("No currency in your saved config (%s), defaulting to USD.\nSet a currency using %s.",
						config.ConfigurationFilePath(),
						ui.PrimaryString("infracost configure set currency CURRENCY"),
					)
					logging.Logger.Warn().Msg(msg)
				}
			case "tls_insecure_skip_verify":
				if ctx.Config.Configuration.TLSInsecureSkipVerify == nil {
					value = ""
				} else {
					value = fmt.Sprintf("%t", *ctx.Config.Configuration.TLSInsecureSkipVerify)
				}

				if value == "" {
					msg := fmt.Sprintf("Skipping TLS verification is not set in your saved config (%s).\nSet it using %s.",
						config.ConfigurationFilePath(),
						ui.PrimaryString("infracost configure set tls_insecure_skip_verify true"),
					)
					logging.Logger.Warn().Msg(msg)
				}
			case "tls_ca_cert_file":
				value = ctx.Config.Configuration.TLSCACertFile

				if value == "" {
					msg := fmt.Sprintf("No CA cert file in your saved config (%s).\nSet a CA certificate using %s.",
						config.ConfigurationFilePath(),
						ui.PrimaryString("infracost configure set tls_ca_cert_file /path/to/ca.crt"),
					)
					logging.Logger.Warn().Msg(msg)
				}
			case "enable_dashboard":
				if ctx.Config.Configuration.EnableDashboard == nil {
					value = ""
				} else {
					value = strconv.FormatBool(*ctx.Config.Configuration.EnableDashboard)
				}
			case "enable_cloud":
				if ctx.Config.Configuration.EnableCloud == nil {
					value = ""
				} else {
					value = strconv.FormatBool(*ctx.Config.Configuration.EnableCloud)
				}
			}

			if value != "" {
				fmt.Println(value)
			}

			return nil
		},
	}

	return cmd
}

func isValidConfigureKey(key string) bool {
	_, ok := supportedConfigureKeys[key]

	return ok
}

func supportedConfigureSettingsOutput(description string) string {
	settings := `
Supported settings:
  - api_key: Infracost API key
  - pricing_api_endpoint: endpoint of the Cloud Pricing API
  - currency: convert output from USD to your preferred currency
  - tls_insecure_skip_verify: skip TLS certificate checks for a self-hosted Cloud Pricing API
  - tls_ca_cert_file: verify certificate of a self-hosted Cloud Pricing API using this CA certificate
`

	return fmt.Sprintf("%s.\n%s", description, settings)
}

func validConfigureKeys() []string {
	keys := make([]string, len(supportedConfigureKeys))

	i := 0
	for k := range supportedConfigureKeys {
		keys[i] = k
		i++
	}

	return keys
}
