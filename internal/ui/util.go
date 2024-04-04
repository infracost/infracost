package ui

import (
	"regexp"
	"strings"
)

func Indent(s, indent string) string {
	lines := make([]string, 0)

	split := strings.Split(s, "\n")

	for i, j := range split {
		if StripColor(j) == "" && i == len(split)-1 {
			lines = append(lines, j)
		} else {
			lines = append(lines, indent+j)
		}
	}
	return strings.Join(lines, "\n")
}

func StripColor(str string) string {
	ansi := "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"
	re := regexp.MustCompile(ansi)
	return re.ReplaceAllString(str, "")
}
