package aws

import (
	"github.com/shopspring/decimal"
)

func strPtr(s string) *string {
	return &s
}

func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}

func int64Ptr(i int64) *int64 {
	return &i
}

func asGiB(i int64) int64 {
	i = i / (1024 * 1024 * 1024)
	if i == 0 {
		return 1
	} else {
		return i
	}
}
