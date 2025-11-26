package google

import (
	"regexp"
	"slices"
	"strings"

	"github.com/shopspring/decimal"
)

var defaultVolumeSize = 10

func intPtr(i int64) *int64 {
	return &i
}

func strPtr(s string) *string {
	return &s
}

// nolint:deadcode,unused
func floatPtr(f float64) *float64 {
	return &f
}

func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}

func isZone(location string) bool {
	if matched, _ := regexp.MatchString(`^\w+-\w+-\w+$`, location); matched {
		return true
	}
	return false
}

func zoneToRegion(zone string) string {
	s := strings.Split(zone, "-")
	return strings.Join(s[:len(s)-1], "-")
}

func contains(a []string, x string) bool {
	return slices.Contains(a, x)
}
