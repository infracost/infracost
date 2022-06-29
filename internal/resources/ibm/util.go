package ibm

import (
	"fmt"
	"regexp"

	"github.com/shopspring/decimal"
)

var (
	vendorName = strPtr("ibm")
	underscore = regexp.MustCompile(`_`)
)

func strPtr(s string) *string {
	return &s
}

// nolint:deadcode,unused
func regexPtr(regex string) *string {
	return strPtr(fmt.Sprintf("/%s/i", regex))
}

func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}

// nolint:deadcode,unused
func floatPtrToDecimalPtr(f *float64) *decimal.Decimal {
	if f == nil {
		return nil
	}
	return decimalPtr(decimal.NewFromFloat(*f))
}

// map to Global Catalog service if
type ibmMetadata struct {
	serviceId string
}
