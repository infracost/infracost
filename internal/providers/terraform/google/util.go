package google

import (
	"strings"

	"github.com/shopspring/decimal"
)

var defaultVolumeSize = 10

func strPtr(s string) *string {
	return &s
}

func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}

func zoneToRegion(zone string) string {
	s := strings.Split(zone, "-")
	return strings.Join(s[:len(s)-1], "-")
}
