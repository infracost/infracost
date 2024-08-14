package output

import (
	"testing"

	"github.com/Rhymond/go-money"

	"github.com/google/go-cmp/cmp"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestFormatCost(t *testing.T) {
	tests := map[string]struct {
		currency string
		val      string
		expected string
	}{
		"rounds large USD":        {currency: "USD", val: "1234.567", expected: "$1,235"}, // format cost rounds large numbers to whole units
		"rounds small USD":        {currency: "USD", val: "1.234567", expected: "$1"},
		"rounds really small USD": {currency: "USD", val: "0.01234567890123456789", expected: "$0.01"},
		"rounds large PAB":        {currency: "PAB", val: "1234.567", expected: "B/.1,235"},
		"rounds small PAB":        {currency: "PAB", val: "1.234567", expected: "B/.1"},
		"rounds really small PAB": {currency: "PAB", val: "0.01234567890123456789", expected: "B/.0.01"},
		"rounds large BSD":        {currency: "BSD", val: "1234.567", expected: "$1,235"},
		"rounds small BSD":        {currency: "BSD", val: "1.234567", expected: "$1"},
		"rounds really small BSD": {currency: "BSD", val: "0.01234567890123456789", expected: "$0.01"},
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
				t.Fatal(diff)
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
		"rounds large BSD":        {currency: "BSD", val: "1234.567", expected: "$1,234.57"},
		"rounds small BSD":        {currency: "BSD", val: "1.234567", expected: "$1.23"},
		"rounds really small BSD": {currency: "BSD", val: "0.01234567890123456789", expected: "$0.01"},
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

			got := FormatCost2DP(tc.currency, val)

			diff := cmp.Diff(tc.expected, got)
			if diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestCurrencyFormatCost(t *testing.T) {
	tests := map[string]struct {
		format   string
		val      string
		expected string
	}{
		"rounds large USD":                               {format: "USD: $1,234.567890", val: "1234.56789", expected: "$1,235"},
		"rounds large USD with space":                    {format: "USD: 1.234,567890 $", val: "1234.56789", expected: "1.235 $"},
		"rounds small USD":                               {format: "USD: $1,234.56", val: "1.234567", expected: "$1"},
		"rounds really small USD":                        {format: "USD: $1,234.56", val: "0.01234567890123456789", expected: "$0.01"},
		"rounds large PAB":                               {format: "PAB: B/.1,234", val: "1234.567", expected: "B/.1,235"},
		"rounds small PAB":                               {format: "PAB: B/.1,234.56", val: "1.234567", expected: "B/.1"},
		"rounds really small PAB":                        {format: "PAB: B/.1,234.56", val: "0.01234567890123456789", expected: "B/.0.01"},
		"rounds small BSD":                               {format: "BSD: $1,234.56", val: "1.234567", expected: "$1"},
		"rounds really small BSD":                        {format: "BSD: $1,234.56", val: "0.01234567890123456789", expected: "$0.01"},
		"handles nil":                                    {format: "USD: $1,234.56", val: "", expected: "-"},
		"handles invalid format with the default format": {format: "USD: $1,23.56,7", val: "1.234567", expected: "$1"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			reset := resetCurrencyFormatFunc(tc.format)
			addCurrencyFormat(tc.format)
			defer reset()

			var val *decimal.Decimal
			if tc.val != "" {
				parsed, err := decimal.NewFromString(tc.val)
				require.NoError(t, err)
				val = &parsed
			}

			currency := tc.format[0:3]
			got := formatCost(currency, val)

			diff := cmp.Diff(tc.expected, got)
			if diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func resetCurrencyFormatFunc(format string) func() {
	code := format[0:3]
	currency := money.GetCurrency(code)
	return func() {
		money.AddCurrency(currency.Code, currency.Grapheme, currency.Template, currency.Decimal, currency.Thousand, currency.Fraction)
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
		"rounds large BSD":        {currency: "BSD", val: "1234.567", expected: "$1,234.57"},
		"rounds small BSD":        {currency: "BSD", val: "1.234567", expected: "$1.23"},
		"rounds really small BSD": {currency: "BSD", val: "0.01234567890123456789", expected: "$0.0123456789"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			val, err := decimal.NewFromString(tc.val)
			require.NoError(t, err)

			got := formatPrice(tc.currency, val)

			diff := cmp.Diff(tc.expected, got)
			if diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
