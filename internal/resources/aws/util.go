package aws

import (
	"math"

	"github.com/shopspring/decimal"
)

func strPtr(s string) *string {
	return &s
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func intPtr(i int64) *int64 {
	return &i
}

func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}

func asGiB(i int64) int64 {
	if i == 0 {
		return 0
	}
	i /= (1024 * 1024 * 1024)
	if i == 0 {
		return 1
	}
	return i
}

func ceil64(f float64) int64 {
	return int64(math.Ceil(f))
}

func stringInSlice(slice []string, s string) bool {
	for _, b := range slice {
		if b == s {
			return true
		}
	}
	return false
}
