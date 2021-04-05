package azure

import (
	"strings"

	"github.com/shopspring/decimal"
)

func strPtr(s string) *string {
	return &s
}

func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}

// A regular expression forbidding a particular substring.
// This function is case insensitive.
func regexMustNotContain(s string) string {
	return "/^((?!" + s + ").)*$/i"
}

// A regular expression where string must contain a particular substring.
// This function is case insensitive.
func regexMustContain(s string) string {
	return "/" + s + "/i"
}

// Parse from Terraform size value to Azure instance type value.
func parseInstanceType(size string) string {
	s := strings.ReplaceAll(size, "Standard_", "")
	s = strings.ReplaceAll(s, "Basic_", "")
	s = strings.ReplaceAll(s, "_", " ")

	if strings.HasPrefix(size, "Basic_") {
		return "Basic " + s
	}
	return s
}
