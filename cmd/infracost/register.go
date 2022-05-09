package main

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
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

			if ctx.Config.Credentials.APIKey != "" {

				isRegenerate = true
				fmt.Printf("You already have an Infracost API key saved in %s. We recommend using your same API key in all environments.\n", config.CredentialsFilePath())

				status, err := promptGenerateNewKey()
				if err != nil {
					return err
				}

				if !status {
					showDocs()
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

			d := apiclient.NewDashboardAPIClient(ctx)

			r, err := d.CreateAPIKey(name, email, ctx.ContextValues())
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
					showDocs()
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
			showDocs()
			return nil
		},
	}
}

func showDocs() {
	fmt.Printf("Follow %s to use Infracost.\n", ui.LinkString("https://infracost.io/docs"))
}

func promptForName() (string, error) {
	name, err := stringPrompt("Name", func(input string) error {
		input = strings.TrimSpace(input)
		if input == "" {
			return errors.New("Please enter a name")
		}
		return nil
	})

	return name, err
}

func promptForEmail() (string, error) {
	email, err := stringPrompt("Email", func(input string) error {
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
	})

	return email, err
}

func promptOverwriteAPIKey() (bool, error) {
	return yesNoPrompt("Would you like to overwrite your existing saved API key")
}

func promptGenerateNewKey() (bool, error) {
	return yesNoPrompt("Would you like to generate a new API key")
}

func stringPrompt(label string, validate ui.ValidateFn) (string, error) {
	if runtime.GOOS == "windows" {
		return ui.StringPrompt(label, validate), nil
	}

	p := promptui.Prompt{
		Label:    label,
		Validate: promptui.ValidateFunc(validate),
	}
	name, err := p.Run()
	name = strings.TrimSpace(name)
	return name, err
}

func yesNoPrompt(label string) (bool, error) {
	if runtime.GOOS == "windows" {
		return ui.YesNoPrompt(label + "?"), nil
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
