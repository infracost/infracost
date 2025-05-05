package config

import (
	"fmt"
	"regexp"
	"strings"
)

type TerraformSourceMap map[string]string

func (s *TerraformSourceMap) Decode(value string) error {
	sourceMap := map[string]string{}
	pairs := strings.Split(value, ",")
	for _, pair := range pairs {
		kvpair := maskAndSplit(pair)

		if len(kvpair) != 2 {
			return fmt.Errorf("invalid map item: %q", pair)
		}
		sourceMap[kvpair[0]] = kvpair[1]
	}

	*s = TerraformSourceMap(sourceMap)
	return nil
}

// maskAndSplit replaces the special characters in the string with a mask
// and then splits the string by the mask.
// This uses the same masking technique as terragrunt for the
// TERRAGRUNT_SOURCE_MAP environment variable.
func maskAndSplit(s string) []string {
	masks := map[string]string{
		"?ref=": "<ref-place-holder>",
	}

	for src, mask := range masks {
		s = strings.ReplaceAll(s, src, mask)
	}

	values := strings.Split(s, "=")

	for i := range values {
		for src, mask := range masks {
			values[i] = strings.ReplaceAll(values[i], mask, src)
		}
	}

	return values
}

// TerraformSourceMapRegexEntry represents a single regex-based source map entry
type TerraformSourceMapRegexEntry struct {
	// Match is the regex pattern to match on the source URL
	Match string `yaml:"match"`
	// Replace is the replacement pattern which can contain capture groups
	Replace string `yaml:"replace"`
	// compiled regex pattern (not exported in YAML)
	pattern *regexp.Regexp
}

// TerraformSourceMapRegex is a slice of regex-based source map entries
type TerraformSourceMapRegex []TerraformSourceMapRegexEntry

func (s *TerraformSourceMapRegex) IsCompiled() bool {
	for _, v := range *s {
		if v.pattern == nil {
			return false
		}
	}

	return true
}

// Compile compiles all regex patterns for faster matching
func (s *TerraformSourceMapRegex) Compile() error {
	if s.IsCompiled() {
		return nil
	}

	for i, entry := range *s {
		pattern, err := regexp.Compile(entry.Match)
		if err != nil {
			return fmt.Errorf("invalid regex pattern in source map entry %d (%s): %w", i, entry.Match, err)
		}
		(*s)[i].pattern = pattern
	}
	return nil
}

// ApplyRegexMapping applies the regex-based mapping to the given source URL
// and returns the mapped URL if there is a match, otherwise returns the original source
func ApplyRegexMapping(sourceMap TerraformSourceMapRegex, source string) (string, error) {
	for _, entry := range sourceMap {
		if entry.pattern == nil {
			return "", fmt.Errorf("regex pattern not compiled for %s", entry.Match)
		}

		if entry.pattern.MatchString(source) {
			// Use regexp to replace with support for capture groups
			result := entry.pattern.ReplaceAllString(source, entry.Replace)
			return result, nil
		}
	}

	// No match found, return the original source
	return source, nil
}
