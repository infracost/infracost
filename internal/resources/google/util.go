package google

import (
	"github.com/shopspring/decimal"
)

func strPtr(s string) *string {
	return &s
}

func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}

func intPtrToDecimalPtr(i *int64) *decimal.Decimal {
	if i == nil {
		return nil
	}
	return decimalPtr(decimal.NewFromInt(*i))
}
