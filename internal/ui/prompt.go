package ui

import (
	"fmt"
	"regexp"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

var yes = regexp.MustCompile(`(?i)ye?s?`)

func CanPrompt() bool {
	return terminal.IsTerminal(syscall.Stdin)
}

func PromptBool(prompt string) bool {
	answer := "n"
	fmt.Printf("%s [y/N]: ", prompt)
	fmt.Scanln(&answer)
	return yes.MatchString(answer)
}
