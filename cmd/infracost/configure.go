package main

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/ui"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var validConfigureKeys = []string{"api_key", "pricing_api_endpoint"}

func configureCmd(ctx *config.RunContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configure",
		Short: "Configure Infracost",
		Long:  "Configure Infracost CLI",
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
		Short: "Set Infracost CLI global configuration",
		Long:  "Set Infracost CLI global configuration",
		Example: `  Set your Infracost API key:

      infracost configure set api_key API_KEY_VALUE

  Set your Infracost Pricing API Endpoint:

      infracost	configure set pricing_api_endpoint https://cloud-pricing-api.internal

  Set your Infracost API key for a specific Pricing API Endpoint:

      infracost configure set api_key API_KEY_VALUE --pricing-api-endpoint https://cloud-pricing-api.internal`,
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

			if cmd.Flags().Changed("pricing-api-endpoint") && key != "api_key" {
				ui.PrintWarning("--pricing-api-endpoint flag is only used when key is api_key")
			}

			if key == "pricing_api_endpoint" {
				if existing, ok := ctx.Config.Credentials[value]; ok {
					existing.Default = true
				} else {
					ctx.Config.Credentials[value] = &config.CredentialsProfileSpec{
						Default: true,
					}
				}

				for k, v := range ctx.Config.Credentials {
					if k != value {
						v.Default = false
					}
				}

				err := ctx.Config.Credentials.Save()
				if err != nil {
					return err
				}
			}

			if key == "api_key" {
				pricingAPIEndpoint, _ := cmd.Flags().GetString("pricing-api-endpoint")
				if pricingAPIEndpoint == "" {
					pricingAPIEndpoint = ctx.Config.PricingAPIEndpoint
				}

				ctx.Config.Credentials[pricingAPIEndpoint] = &config.CredentialsProfileSpec{
					APIKey:  value,
					Default: true,
				}

				for k, v := range ctx.Config.Credentials {
					if k != pricingAPIEndpoint {
						v.Default = false
					}
				}

				err := ctx.Config.Credentials.Save()
				if err != nil {
					return err
				}
			}

			return nil
		},
	}

	cmd.Flags().String("pricing-api-endpoint", "", "The Pricing API Endpoint. Only applicable when key is api_key")

	return cmd
}

func configureGetCmd(ctx *config.RunContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Get Infracost CLI global configuration",
		Long:  "Get Infracost CLI global configuration",
		Example: `  Get your saved Infracost API key:

      infracost configure get api_key

  Get your saved Infracost Pricing API Endpoint:

      infracost	configure get pricing_api_endpoint

  Get your saved Infracost API key for a specific Pricing API Endpoint:

      infracost configure get api_key --pricing-api-endpoint https://cloud-pricing-api.internal`,
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

			if cmd.Flags().Changed("pricing-api-endpoint") && key != "api_key" {
				ui.PrintWarning("--pricing-api-endpoint flag is only used when key is api_key")
			}

			if key == "pricing_api_endpoint" {
				value = ctx.Config.Credentials.GetDefaultPricingAPIEndpoint()

				if value == "" {
					ui.PrintWarning("Could not find any Pricing API Endpoint in your saved config")
				}
			}

			if key == "api_key" {
				pricingAPIEndpoint, _ := cmd.Flags().GetString("pricing-api-endpoint")
				if pricingAPIEndpoint == "" {
					pricingAPIEndpoint = ctx.Config.Credentials.GetDefaultPricingAPIEndpoint()
				}

				profile := ctx.Config.Credentials[pricingAPIEndpoint]

				if profile != nil {
					value = profile.APIKey
				} else {
					if pricingAPIEndpoint != ctx.Config.DefaultPricingAPIEndpoint {
						ui.PrintWarningf("Could not find an API key for Pricing API Endpoint %s in your saved config", pricingAPIEndpoint)
					} else {
						ui.PrintWarning("Could not find an API key in your saved config")
					}
				}
			}

			if value != "" {
				fmt.Println(value)
			}

			return nil
		},
	}

	cmd.Flags().String("pricing-api-endpoint", "", "The Pricing API Endpoint. Only applicable when key is api_key")

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
