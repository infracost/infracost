package aws

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestConvertResourceString(t *testing.T) {
	half := decimal.NewFromFloat(0.5)
	one := decimal.NewFromInt(1)
	two := decimal.NewFromInt(2)

	tests := []struct {
		inputStr string
		expected decimal.Decimal
	}{
		{"1GB", one},
		{"1gb", one},
		{" 1 Gb ", one}, // mixed case and pre/middle/post whitespace
		{"0.5 GB", half},
		{".5 GB", half},
		{"1VCPU", one},
		{"1vcpu", one},
		{" 1 vCPU ", one}, // mixed case and pre/middle/post whitespace
		{"1024", one},
		{" 1024 ", one},
		{"512", half},
		{"2048", two},
	}

	for _, test := range tests {
		actual := convertResourceString(test.inputStr)
		assert.Equal(t, test.expected.String(), actual.String())
	}
}
