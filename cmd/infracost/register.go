package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/infracost/infracost/internal/apiclient"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/ui"
)

func registerCmd(ctx *config.RunContext) *cobra.Command {
	return &cobra.Command{
		Use:   "register",
		Short: "Register for a free Infracost API key",
		Long:  "Register for a free Infracost API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			var isRegenerate bool
			var ciInterest bool

			if ctx.Config.Credentials.APIKey != "" {

				isRegenerate = true
				fmt.Printf("You already have an Infracost API key saved in %s. We recommend using your same API key in all environments.\n", config.CredentialsFilePath())

				status, err := promptGenerateNewKey()
				if err != nil {
					return err
				}

				if !status {
					ciInterest, err := promptForCIDocs(false)
					if err != nil {
						// user cancelled
						return nil
					}

					if ciInterest {
						fmt.Println("Add cost estimates to your pull requests: " + ui.LinkString("https://infracost.io/cicd"))
						return nil
					}

					fmt.Printf("You can now run %s and point to your Terraform directory or JSON plan file.\n", ui.PrimaryString("infracost breakdown --path=..."))
					return nil
				}

				ciInterest, err = promptForCIDocs(isRegenerate)
				if err != nil {
					// user cancelled
					return nil
				}

				fmt.Println()
			}

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

			// prompt for the ci docs after user email,name only if not regenerating
			if !isRegenerate {
				ciInterest, err = promptForCIDocs(isRegenerate)
				if err != nil {
					// user cancelled
					return nil
				}
			}

			d := apiclient.NewDashboardAPIClient(ctx)

			r, err := d.CreateAPIKey(name, email, ciInterest)
			if err != nil {
				return err
			}

			if r.Error != "" {
				fmt.Fprintln(os.Stderr, "")
				ui.PrintErrorf(cmd.ErrOrStderr(), "There was an error requesting an API key\n%s\nPlease contact hello@infracost.io if you continue to have issues.", r.Error)
				return nil
			}

			fmt.Printf("\nThanks %s! Your API key is: %s\n", name, r.APIKey)

			if isRegenerate {
				fmt.Println()
				confirm, err := promptOverwriteAPIKey()
				if err != nil {
					return err
				}

				if !confirm {
					fmt.Printf("%s\n%s %s %s",
						"Setting the INFRACOST_API_KEY environment variable overrides the key from credentials.yml.",
						"You can now run",
						ui.PrimaryString("infracost breakdown --path=..."),
						"and point to your Terraform directory or plan JSON file.\n",
					)

					return nil
				}
			}

			ctx.Config.Credentials.APIKey = r.APIKey
			ctx.Config.Credentials.PricingAPIEndpoint = ctx.Config.PricingAPIEndpoint

			err = ctx.Config.Credentials.Save()
			if err != nil {
				return err
			}

			fmt.Printf("This was saved to %s\n\n", config.CredentialsFilePath())
			if ciInterest {
				fmt.Printf("You can now add cost estimates to your pull requests: %s\n", ui.LinkString("https://infracost.io/cicd"))
				return nil
			}

			if isRegenerate {
				fmt.Printf("You can now run %s and point to your Terraform directory or JSON plan file.\n", ui.PrimaryString("infracost breakdown --path=..."))
				return nil
			}

			fmt.Printf("You can now run %s to see how to use the CLI\n", ui.PrimaryString("infracost breakdown --help"))
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

func promptForCIDocs(isRegenerate bool) (bool, error) {
	label := "Would you like to see our CI/CD integration docs?"
	if isRegenerate {
		label = "Do you plan to use this API key in CI?"
	}
	p := promptui.Prompt{
		Label:     label,
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

func promptGenerateNewKey() (bool, error) {
	p := promptui.Prompt{
		Label:     "Would you like to generate a new API key",
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
