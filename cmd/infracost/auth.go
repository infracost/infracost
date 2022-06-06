package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/infracost/infracost/internal/apiclient"
	"github.com/infracost/infracost/internal/config"
)

func authCmd(ctx *config.RunContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Get a free API key, or login to your existing account",
		Long:  "Get a free API key, or login to your existing account",
		Example: `  Get a free API key, or login to your existing account:

      infracost auth login

  Manually set the API key that your CLI should use. The API key can be retrieved from your account:

      infracost configure set api_key MY_API_KEY`,
		ValidArgs: []string{"--", "-"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmds := []*cobra.Command{authLoginCmd(ctx)}
	cmd.AddCommand(cmds...)

	return cmd
}

func authLoginCmd(ctx *config.RunContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate the CLI with your Infracost account",
		Long:  "Authenticate the CLI with your Infracost account",
		Example: `  Login:

      infracost auth login`,
		ValidArgs: []string{"--", "-"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.Println("We're redirecting you to our login page, please complete that,\nand return here to continue using Infracost.")

			auth := apiclient.AuthClient{Host: ctx.Config.DashboardEndpoint}
			apiKey, err := auth.Login()
			if err != nil {
				cmd.Println(err)
				return err
			}

			ctx.Config.Credentials.APIKey = apiKey
			ctx.Config.Credentials.PricingAPIEndpoint = ctx.Config.PricingAPIEndpoint

			err = ctx.Config.Credentials.Save()
			if err != nil {
				return err
			}

			fmt.Printf("The API key was saved to %s\n", config.CredentialsFilePath())
			cmd.Println("\nYour account has been authenticated. Run Infracost on your Terraform project by running:")
			cmd.Printf("\n  infracost breakdown --path=.\n\n")

			return nil
		},
	}

	return cmd
}
