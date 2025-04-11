package util

import "regexp"

var redactionRegex = regexp.MustCompile(`(?im)://(.+):(.+)@`)

// redact source removes username:password credentials from sources
func RedactUrl(source string) string {
	return redactionRegex.ReplaceAllString(source, "://****:****@")
}

func RedactUrlPtr(moduleLocation *string) *string {
	if moduleLocation == nil {
		return nil
	}
	redacted := RedactUrl(*moduleLocation)
	return &redacted
}
