package config

import (
	"fmt"
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
