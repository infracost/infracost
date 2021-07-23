package aws

import "github.com/shopspring/decimal"

var defaultVolumeSize = 8

func intPtr(i int64) *int64 {
	return &i
}

func strPtr(s string) *string {
	return &s
}

func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}

func stringInSlice(slice []string, s string) bool {
	for _, b := range slice {
		if b == s {
			return true
		}
	}
	return false
}
