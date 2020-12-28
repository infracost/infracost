package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/fatih/color"
	"github.com/infracost/infracost/internal/config"
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

type submitFeedbackResp struct {
	Error string `json:"error"`
}

func feedbackCmd() *cli.Command {
	return &cli.Command{
		Name:  "feedback",
		Usage: "Submit feedback directly to the Infracost team",
		Action: func(c *cli.Context) error {
			fmt.Println("Please enter any feedback you have for us.")
			fmt.Println("If you'd like to contact us directly, please email us at hello@infracost.io")
			fmt.Println("Press ENTER to submit feedback.")

			feedback, err := promptForFeedback()
			if err != nil {
				// user cancelled
				return nil
			}

			r, err := submitFeedback(feedback)
			if err != nil || r.Error != "" {
				color.Red("There was an error submitting your feedback:\n%s\n", r.Error)
				fmt.Println("Please contact hello@infracost.io if you continue to have issues.")
				return nil
			}

			color.Green("\nThank you for submitting your feedback\n")

			return nil
		},
	}
}

func promptForFeedback() (string, error) {
	p := promptui.Prompt{
		Label: "Feedback",
		Validate: func(input string) error {
			input = strings.TrimSpace(input)
			if input == "" {
				return errors.New("Please enter your feedback")
			}
			return nil
		},
	}
	return p.Run()
}

func submitFeedback(feedback string) (*submitFeedbackResp, error) {
	url := fmt.Sprintf("%s/feedback-cli", config.Config.DashboardAPIEndpoint)
	d := map[string]string{"feedback": feedback}

	j, err := json.Marshal(d)
	if err != nil {
		return nil, errors.Wrap(err, "Error submitting feedback")
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(j))
	if err != nil {
		return nil, errors.Wrap(err, "Error submitting feedback")
	}

	config.AddAuthHeaders(req)

	client := http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return nil, errors.Wrap(err, "Error sending feedback request")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Invalid response from API")
	}

	var r submitFeedbackResp

	err = json.Unmarshal(body, &r)
	if err != nil {
		return nil, errors.Wrap(err, "Invalid response from API")
	}

	return &r, nil
}
