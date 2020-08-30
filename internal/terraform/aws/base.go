package aws

import "github.com/shopspring/decimal"

var DefaultVolumeSize = 8

func strPtr(s string) *string {
	return &s
}

func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}
