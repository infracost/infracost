package azure

import (
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
func regexDoesNotContain(s string) string {
	return "/^((?!" + s + ").)*$/i"
}

// A regular expression where string must contain a particular substring.
// This function is case insensitive.
func regexMustContain(s string) string {
	return "/" + s + "/i"
}
