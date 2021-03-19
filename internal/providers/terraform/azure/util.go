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

// Parse from Terraform Size value to Azure Pricing SKU Name value.
func parseVirtualMachineSizeSKU(s string) string {
	sku := strings.ReplaceAll(s, "Standard_", "")
	sku = strings.ReplaceAll(sku, "Basic_", "")
	sku = strings.ReplaceAll(sku, "_", " ")
	return sku
}
