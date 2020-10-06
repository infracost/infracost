package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/infracost/infracost/pkg/config"
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

type createAPIKeyResponse struct {
	Error string `json:"error"`
}

func registerCmd() *cli.Command {
	return &cli.Command{
		Name: "register",
		Action: func(c *cli.Context) error {
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

			r, err := createApiKey(name, email)
			if err != nil {
				return err
			}

			if r.Error != "" {
				color.Red("There was an error requesting an API key:\n%s\n", r.Error)
				fmt.Println("Please contact hello@infracost.io if you continue to have issues.")
				return nil
			}

			color.Green("Thank you %s!", name)
			color.Green("Your API key has been sent to %s\n", email)
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

func createApiKey(name string, email string) (*createAPIKeyResponse, error) {
	url := fmt.Sprintf("%s/apiKeys?source=cli-register", config.Config.DashboardAPIEndpoint)
	d := map[string]string{"name": name, "email": email}
	j, err := json.Marshal(d)
	if err != nil {
		return nil, errors.Wrap(err, "Error generating API key request")
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(j)))
	if err != nil {
		return nil, errors.Wrap(err, "Error generating API key request")
	}

	req.Header.Set("content-type", "application/json")
	req.Header.Set("User-Agent", config.GetUserAgent())

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "Error sending API key request")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Invalid reponse from API")
	}
	var r createAPIKeyResponse
	err = json.Unmarshal(body, &r)
	if err != nil {
		return nil, errors.Wrap(err, "Invalid reponse from API")
	}

	return &r, nil
}
