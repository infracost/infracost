package output

import (
	"github.com/google/go-cmp/cmp"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFormatCost(t *testing.T) {
	tests := map[string]struct {
		currency string
		val      string
		expected string
	}{
		"rounds large USD":        {currency: "USD", val: "1234.567", expected: "$1,235"}, // format cost rounds large numbers to whole units
		"rounds small USD":        {currency: "USD", val: "1.234567", expected: "$1.23"},
		"rounds really small USD": {currency: "USD", val: "0.01234567890123456789", expected: "$0.01"},
		"rounds large PAB":        {currency: "PAB", val: "1234.567", expected: "B/.1,235"},
		"rounds small PAB":        {currency: "PAB", val: "1.234567", expected: "B/.1.23"},
		"rounds really small PAB": {currency: "PAB", val: "0.01234567890123456789", expected: "B/.0.01"},
		"rounds large BSD":        {currency: "BSD", val: "1234.567", expected: "$1,235 BSD"},
		"rounds small BSD":        {currency: "BSD", val: "1.234567", expected: "$1.23 BSD"},
		"rounds really small BSD": {currency: "BSD", val: "0.01234567890123456789", expected: "$0.01 BSD"},
		"handles nil":             {currency: "USD", val: "", expected: "-"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var val *decimal.Decimal
			if tc.val != "" {
				parsed, err := decimal.NewFromString(tc.val)
				require.NoError(t, err)
				val = &parsed
			}

			got := formatCost(tc.currency, val)

			diff := cmp.Diff(tc.expected, got)
			if diff != "" {
				t.Fatalf(diff)
			}
		})
	}
}

func TestFormatCost2DP(t *testing.T) {
	tests := map[string]struct {
		currency string
		val      string
		expected string
	}{
		"rounds large USD":        {currency: "USD", val: "1234.567", expected: "$1,234.57"},
		"rounds small USD":        {currency: "USD", val: "1.234567", expected: "$1.23"},
		"rounds really small USD": {currency: "USD", val: "0.01234567890123456789", expected: "$0.01"},
		"rounds large PAB":        {currency: "PAB", val: "1234.567", expected: "B/.1,234.57"},
		"rounds small PAB":        {currency: "PAB", val: "1.234567", expected: "B/.1.23"},
		"rounds really small PAB": {currency: "PAB", val: "0.01234567890123456789", expected: "B/.0.01"},
		"rounds large BSD":        {currency: "BSD", val: "1234.567", expected: "$1,234.57 BSD"},
		"rounds small BSD":        {currency: "BSD", val: "1.234567", expected: "$1.23 BSD"},
		"rounds really small BSD": {currency: "BSD", val: "0.01234567890123456789", expected: "$0.01 BSD"},
		"handles nil":             {currency: "USD", val: "", expected: "-"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var val *decimal.Decimal
			if tc.val != "" {
				parsed, err := decimal.NewFromString(tc.val)
				require.NoError(t, err)
				val = &parsed
			}

			got := formatCost2DP(tc.currency, val)

			diff := cmp.Diff(tc.expected, got)
			if diff != "" {
				t.Fatalf(diff)
			}
		})
	}
}

func TestFormatPrice(t *testing.T) {
	tests := map[string]struct {
		currency string
		val      string
		expected string
	}{
		"rounds large USD":        {currency: "USD", val: "1234.567", expected: "$1,234.57"},
		"rounds small USD":        {currency: "USD", val: "1.234567", expected: "$1.23"},
		"rounds really small USD": {currency: "USD", val: "0.01234567890123456789", expected: "$0.0123456789"},
		"rounds large PAB":        {currency: "PAB", val: "1234.567", expected: "B/.1,234.57"},
		"rounds small PAB":        {currency: "PAB", val: "1.234567", expected: "B/.1.23"},
		"rounds really small PAB": {currency: "PAB", val: "0.01234567890123456789", expected: "B/.0.0123456789"},
		"rounds large BSD":        {currency: "BSD", val: "1234.567", expected: "$1,234.57 BSD"},
		"rounds small BSD":        {currency: "BSD", val: "1.234567", expected: "$1.23 BSD"},
		"rounds really small BSD": {currency: "BSD", val: "0.01234567890123456789", expected: "$0.0123456789 BSD"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			val, err := decimal.NewFromString(tc.val)
			require.NoError(t, err)

			got := formatPrice(tc.currency, val)

			diff := cmp.Diff(tc.expected, got)
			if diff != "" {
				t.Fatalf(diff)
			}
		})
	}
}
