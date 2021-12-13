package main

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/ui"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var validConfigureKeys = []string{
	"api_key",
	"pricing_api_endpoint",
	"tls_insecure_skip_verify",
	"tls_ca_cert_file",
	"currency",
}

func configureCmd(ctx *config.RunContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configure",
		Short: "Display or change global configuration",
		Long: `Display or change global configuration.

Supported settings:
  - api_key: Infracost API key
  - pricing_api_endpoint: endpoint of the Cloud Pricing API
  - tls_insecure_skip_verify: skip TLS certificate checks for a self-hosted Cloud Pricing API
  - tls_ca_cert_file: verify certificate of a self-hosted Cloud Pricing API using this CA certificate
  - currency: convert output from USD to your preferred currency
`,
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
		Long: `Set global configuration.

Supported settings:
  - api_key: Infracost API key
  - pricing_api_endpoint: endpoint of the Cloud Pricing API
  - tls_insecure_skip_verify: skip TLS certificate checks for a self-hosted Cloud Pricing API
  - tls_ca_cert_file: verify certificate of a self-hosted Cloud Pricing API using this CA certificate
  - currency: convert output from USD to your preferred currency
`,
		Example: `  Set your Infracost API key:

      infracost configure set api_key MY_API_KEY

  Set your Cloud Pricing API endpoint:

      infracost	configure set pricing_api_endpoint https://cloud-pricing-api

  Set your preferred currency code (ISO 4217):

      infracost	configure set currency EUR`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 2 {
				return errors.New("Too many arguments")
			}

			if len(args) < 2 {
				return errors.New("You must specify a key and a value")
			}

			if !isValidConfigureKey(args[0]) {
				return fmt.Errorf("Invalid key, valid keys are: %s", strings.Join(validConfigureKeys, ", "))
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
				ctx.Config().Credentials.PricingAPIEndpoint = value
				saveCredentials = true
			case "api_key":
				ctx.Config().Credentials.APIKey = value
				saveCredentials = true
			case "tls_insecure_skip_verify":
				var b bool
				if value == "true" {
					b = true
				} else if value == "false" {
					b = false
				} else {
					return errors.New("Invalid value, must be true or false")
				}
				ctx.Config().Configuration.TLSInsecureSkipVerify = &b
				saveConfiguration = true
			case "tls_ca_cert_file":
				ctx.Config().Configuration.TLSCACertFile = value
				saveConfiguration = true
			case "currency":
				ctx.Config().Configuration.Currency = value
				saveConfiguration = true
			}

			if saveCredentials {
				err := ctx.Config().Credentials.Save()
				if err != nil {
					return err
				}
			}

			if saveConfiguration {
				err := ctx.Config().Configuration.Save()
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
		Long: `Get global configuration.

Supported settings:
  - api_key: Infracost API key
  - pricing_api_endpoint: endpoint of the Cloud Pricing API
  - tls_insecure_skip_verify: skip TLS certificate checks for a self-hosted Cloud Pricing API
  - tls_ca_cert_file: verify certificate of a self-hosted Cloud Pricing API using this CA certificate
  - currency: convert output from USD to your preferred currency
`,
		Example: `  Get your saved Infracost API key:

      infracost configure get api_key

  Get your saved Cloud Pricing API endpoint:

      infracost	configure get pricing_api_endpoint

  Get your preferred currency:

      infracost	configure get currency`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return errors.New("Too many arguments")
			}

			if len(args) < 1 {
				return errors.New("You must specify a key")
			}

			if !isValidConfigureKey(args[0]) {
				return fmt.Errorf("Invalid key, valid keys are: %s", strings.Join(validConfigureKeys, ", "))
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			var value string

			if key == "pricing_api_endpoint" {
				value = ctx.Config().Credentials.PricingAPIEndpoint

				if value == "" {
					msg := fmt.Sprintf("No Cloud Pricing API endpoint in your saved config (%s).\nSet an API key using %s.",
						config.CredentialsFilePath(),
						ui.PrimaryString("infracost configure set pricing_api_endpoint https://cloud-pricing-api"),
					)
					ui.PrintWarning(cmd.ErrOrStderr(), msg)

				}
			} else if key == "api_key" {
				value = ctx.Config().Credentials.APIKey

				if value == "" {
					msg := fmt.Sprintf("No API key in your saved config (%s).\nSet an API key using %s.",
						config.CredentialsFilePath(),
						ui.PrimaryString("infracost configure set api_key MY_API_KEY"),
					)
					ui.PrintWarning(cmd.ErrOrStderr(), msg)
				}
			} else if key == "tls_insecure_skip_verify" {
				if ctx.Config().Configuration.TLSInsecureSkipVerify == nil {
					value = ""
				} else {
					value = fmt.Sprintf("%t", *ctx.Config().Configuration.TLSInsecureSkipVerify)
				}

				if value == "" {
					msg := fmt.Sprintf("Skipping TLS verification is not set in your saved config (%s).\nSet it using %s.",
						config.ConfigurationFilePath(),
						ui.PrimaryString("infracost configure set tls_insecure_skip_verify true"),
					)
					ui.PrintWarning(cmd.ErrOrStderr(), msg)
				}
			} else if key == "tls_ca_cert_file" {
				value = ctx.Config().Configuration.TLSCACertFile

				if value == "" {
					msg := fmt.Sprintf("No CA cert file in your saved config (%s).\nSet a CA certificate using %s.",
						config.ConfigurationFilePath(),
						ui.PrimaryString("infracost configure set tls_ca_cert_file /path/to/ca.crt"),
					)
					ui.PrintWarning(cmd.ErrOrStderr(), msg)
				}
			} else if key == "currency" {
				value = ctx.Config().Configuration.Currency

				if value == "" {
					msg := fmt.Sprintf("No currency in your saved config (%s), defaulting to USD.\nSet a currency using %s.",
						config.ConfigurationFilePath(),
						ui.PrimaryString("infracost configure set currency CURRENCY"),
					)
					ui.PrintWarning(cmd.ErrOrStderr(), msg)
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
	for _, validKey := range validConfigureKeys {
		if key == validKey {
			return true
		}
	}

	return false
}
