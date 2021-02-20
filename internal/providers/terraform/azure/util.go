package azure

import (
	"github.com/shopspring/decimal"
)

var defaultVolumeSize = 10

func strPtr(s string) *string {
	return &s
}

func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}
