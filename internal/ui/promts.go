package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ValidateFn represents a validation function type for prompts. Function of
// this type should accept a string input and return an error if validation fails;
// otherwise return nil.
type ValidateFn func(input string) error

// StringPrompt provides a single line for user input. It accepts an optional
// validation function.
func StringPrompt(label string, validate ValidateFn) string {
	input := ""
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Fprint(os.Stdout, label+": ")
		input, _ = reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if validate == nil {
			return input
		}

		err := validate(input)
		if err == nil {
			break
		}

		fmt.Fprintln(os.Stderr, err)
	}

	return input
}

// YesNoPrompt provides a yes/no user input. "No" is a default answer if left
// empty.
func YesNoPrompt(label string) bool {
	choices := "y/N"
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Fprintf(os.Stdout, "%s [%s] ", label, choices)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			return false
		}

		switch strings.ToLower(input) {
		case "y", "yes":
			return true
		case "n", "no":
			return false
		}
	}
}
