package ibm

import (
	"fmt"

	"github.com/shopspring/decimal"
)

func strPtr(s string) *string {
	return &s
}

func regexPtr(regex string) *string {
	return strPtr(fmt.Sprintf("/%s/i", regex))
}

func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}
