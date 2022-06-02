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
		Short: "Authenticate with Infracost",
		Long:  "Authenticate with Infracost",
		Example: `  Login:

      infracost auth login

  Manually set the API key that your CLI should use:

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

			auth := apiclient.AuthClient{Host: ctx.Config.DashboardAPIEndpoint}
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
