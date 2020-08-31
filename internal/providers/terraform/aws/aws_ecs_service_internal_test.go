package aws

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/shopspring/decimal"
)

func TestConvertResourceString(t *testing.T) {
	half := decimal.NewFromFloat(0.5)
	one := decimal.NewFromInt(1)
	two := decimal.NewFromInt(2)

	tests := map[string]decimal.Decimal{
		"1GB":      one,
		"1gb":      one,
		" 1 Gb ":   one, // mixed case and pre/middle/post whitespac
		"0.5 GB":   half,
		".5 GB":    half,
		"1VCPU":    one,
		"1vcpu":    one,
		" 1 vCPU ": one, // mixed case and pre/middle/post whitespac
		"1024":     one,
		" 1024 ":   one,
		"512":      half,
		"2048":     two,
	}

	for str, expected := range tests {
		actual := convertResourceString(str)
		if !cmp.Equal(actual, expected) {
			t.Errorf("Conversion of '%s' failed, got: %s, expected: %s", str, actual, expected)
		}
	}
}
