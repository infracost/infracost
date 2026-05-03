package sakura

import "github.com/shopspring/decimal"

func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}

func strPtr(s string) *string {
	return &s
}
