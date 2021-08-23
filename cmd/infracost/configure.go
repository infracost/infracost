package main

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/ui"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var validConfigureKeys = []string{"api_key", "pricing_api_endpoint", "currency"}

func configureCmd(ctx *config.RunContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configure",
		Short: "Display or change global configuration",
		Long: `Display or change global configuration.

Supported settings:
  - api_key: Infracost API key
  - pricing_api_endpoint: endpoint of the Cloud Pricing API
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
  - currency: convert output from USD to your preferred currency
`,
		Example: `  Set your Infracost API key:

      infracost configure set api_key MY_API_KEY

  Set your Cloud Pricing API endpoint:

      infracost	configure set pricing_api_endpoint https://cloud-pricing-api

  Set your preferred currency:

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

			if key == "pricing_api_endpoint" {
				ctx.Config.Credentials.PricingAPIEndpoint = value

				err := ctx.Config.Credentials.Save()
				if err != nil {
					return err
				}
			}

			if key == "api_key" {
				ctx.Config.Credentials.APIKey = value

				err := ctx.Config.Credentials.Save()
				if err != nil {
					return err
				}
			}

			if key == "currency" {
				ctx.Config.Configuration.Currency = value

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
		Long: `Get global configuration.

Supported settings:
  - api_key: Infracost API key
  - pricing_api_endpoint: endpoint of the Cloud Pricing API
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
				value = ctx.Config.Credentials.PricingAPIEndpoint

				if value == "" {
					msg := fmt.Sprintf("No Cloud Pricing API endpoint in your saved config (%s).\nSet an API key using %s.",
						config.CredentialsFilePath(),
						ui.PrimaryString("infracost configure set pricing_api_endpoint https://cloud-pricing-api"),
					)
					ui.PrintWarning(msg)

				}
			} else if key == "api_key" {
				value = ctx.Config.Credentials.APIKey

				if value == "" {
					msg := fmt.Sprintf("No API key in your saved config (%s).\nSet an API key using %s.",
						config.CredentialsFilePath(),
						ui.PrimaryString("infracost configure set api_key MY_API_KEY"),
					)
					ui.PrintWarning(msg)
				}
			} else if key == "currency" {
				value = ctx.Config.Configuration.Currency

				if value == "" {
					msg := fmt.Sprintf("No currency in your saved config (%s), defaulting to USD.\nSet a currency using %s.",
						config.CredentialsFilePath(),
						ui.PrimaryString("infracost configure set currency CURRENCY"),
					)
					ui.PrintWarning(msg)
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
