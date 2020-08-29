package aws

import (
	"github.com/shopspring/decimal"
	"testing"
)

func TestConvertResourceString(t *testing.T) {
	half := decimal.NewFromFloat(0.5)
	one := decimal.NewFromInt(1)
	two := decimal.NewFromInt(2)

	tests := []struct {
		resourceString  string
		resourceDecimal decimal.Decimal
	}{
		{"1GB", one},
		{"1gb", one},
		{" 1 Gb ", one}, // mixed case and pre/middle/post whitespaces
		{"0.5 GB", half},
		{".5 GB", half},
		{"1VCPU", one},
		{"1vcpu", one},
		{" 1 vCPU ", one}, // mixed case and pre/middle/post whitespaces
		{"1024", one},
		{" 1024 ", one},
		{"512", half},
		{"2048", two},
	}

	for _, test := range tests {
		resourceDecimal := convertResourceString(test.resourceString)
		if resourceDecimal.Cmp(test.resourceDecimal) != 0 {
			t.Errorf("Conversion of '%s' failed, got: %s, expected: %s", test.resourceString, resourceDecimal, test.resourceDecimal)
		}
	}
}
