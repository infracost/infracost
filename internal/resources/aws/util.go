package aws

import "github.com/shopspring/decimal"

func strPtr(s string) *string {
	return &s
}

func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}

func floatPtr(f float64) *float64 {
	return &f
}
