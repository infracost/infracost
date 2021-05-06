package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/infracost/infracost/internal/apiclient"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/ui"
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func registerCmd(ctx *config.RunContext) *cobra.Command {
	return &cobra.Command{
		Use:   "register",
		Short: "Register for a free Infracost API key",
		Long:  "Register for a free Infracost API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Please enter your name and email address to get an API key.")
			fmt.Println("See our FAQ (https://www.infracost.io/docs/faq) for more details.")

			name, err := promptForName()
			if err != nil {
				// user cancelled
				return nil
			}

			email, err := promptForEmail()
			if err != nil {
				// user cancelled
				return nil
			}

			d := apiclient.NewDashboardAPIClient(ctx)
			r, err := d.CreateAPIKey(name, email)
			if err != nil {
				return err
			}

			if r.Error != "" {
				fmt.Fprintln(os.Stderr, "")
				ui.PrintErrorf("There was an error requesting an API key\n%s\nPlease contact hello@infracost.io if you continue to have issues.", r.Error)
				return nil
			}

			fmt.Printf("\nThank you %s!\nYour API key is: %s\n", name, r.APIKey)

			msg := fmt.Sprintf("%s\nYou can now run %s and point to your Terraform directory or JSON/plan file.",
				fmt.Sprintf("Your API key has been saved to %s", config.CredentialsFilePath()),
				ui.PrimaryString("infracost breakdown --path=..."),
			)

			saveAPIKey := true

			if _, ok := ctx.Config.Credentials[ctx.Config.PricingAPIEndpoint]; ok {
				fmt.Printf("\nYou already have an Infracost API key saved in %s\n", config.CredentialsFilePath())
				confirm, err := promptOverwriteAPIKey()
				if err != nil {
					return err
				}

				if !confirm {
					saveAPIKey = false

					msg = fmt.Sprintf("%s\n%s %s %s",
						"Setting the INFRACOST_API_KEY environment variable overrides the key from credentials.yml.",
						"You can now run",
						ui.PrimaryString("infracost breakdown --path=..."),
						"and point to your Terraform directory or JSON/plan file.",
					)
				}
			}

			if saveAPIKey {
				ctx.Config.Credentials[ctx.Config.PricingAPIEndpoint] = config.CredentialsProfileSpec{
					APIKey: r.APIKey,
				}

				err = ctx.Config.Credentials.Save()
				if err != nil {
					return err
				}
			}

			fmt.Println("")
			ui.PrintSuccess(msg)

			return nil
		},
	}
}

func promptForName() (string, error) {
	p := promptui.Prompt{
		Label: "Name",
		Validate: func(input string) error {
			input = strings.TrimSpace(input)
			if input == "" {
				return errors.New("Please enter a name")
			}
			return nil
		},
	}
	name, err := p.Run()
	name = strings.TrimSpace(name)

	return name, err
}

func promptForEmail() (string, error) {
	p := promptui.Prompt{
		Label: "Email",
		Validate: func(input string) error {
			input = strings.TrimSpace(input)
			if input == "" {
				return errors.New("Please enter an email")
			}
			match, err := regexp.MatchString(".+@.+", input)
			if err != nil {
				return errors.New("Unable to validate email")
			}
			if !match {
				return errors.New("Please enter a valid email")
			}
			return nil
		},
	}
	email, err := p.Run()
	email = strings.TrimSpace(email)

	return email, err
}

func promptOverwriteAPIKey() (bool, error) {
	p := promptui.Prompt{
		Label:     "Would you like to overwrite your existing saved API key",
		IsConfirm: true,
	}

	_, err := p.Run()
	if err != nil {
		if errors.Is(err, promptui.ErrAbort) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
