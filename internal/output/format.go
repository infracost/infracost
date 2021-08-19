package output

import (
	"github.com/Rhymond/go-money"
	"github.com/dustin/go-humanize"
	"github.com/shopspring/decimal"
	"math"
)

var roundCostsAbove = 100

func formatQuantity(q *decimal.Decimal) string {
	if q == nil {
		return "-"
	}
	f, _ := q.Float64()
	return humanize.CommafWithDigits(f, 4)
}

func formatCost(currency string, d *decimal.Decimal) string {
	if d == nil {
		return "-"
	}
	if d.GreaterThanOrEqual(decimal.NewFromInt(int64(roundCostsAbove))) {
		return formatWholeDecimalCurrency(currency, *d)
	}
	return formatRoundedDecimalCurrency(currency, *d)
}

func formatCost2DP(currency string, d *decimal.Decimal) string {
	if d == nil {
		return "-"
	}
	return formatRoundedDecimalCurrency(currency, *d)
}

func formatPrice(currency string, d decimal.Decimal) string {
	if d.LessThan(decimal.NewFromFloat(0.1)) {
		return formatFullDecimalCurrency(currency, d)
	}
	return formatRoundedDecimalCurrency(currency, d)
}

func formatFullDecimalCurrency(currency string, d decimal.Decimal) string {
	formatter := getCurrencyFormatter(currency)
	scaledInt := decimalToScaledInt(d, formatter.Fraction, 10)
	formatter.Fraction = scaledInt.FractionLength
	return formatter.Format(scaledInt.Number)
}

func formatRoundedDecimalCurrency(currency string, d decimal.Decimal) string {
	formatter := getCurrencyFormatter(currency)

	scaledInt := decimalToScaledInt(d, formatter.Fraction, formatter.Fraction)
	formatter.Fraction = scaledInt.FractionLength
	return formatter.Format(scaledInt.Number)
}

func formatWholeDecimalCurrency(currency string, d decimal.Decimal) string {
	formatter := getCurrencyFormatter(currency)

	scaledInt := decimalToScaledInt(d, 0, 0)
	formatter.Fraction = scaledInt.FractionLength
	return formatter.Format(scaledInt.Number)
}

type scaledInt64 struct {
	Number         int64
	FractionLength int
}

// Convert a decimal to a "scaled int" that can be used with the money.Formatter.
// This is a bit funny since decimal.Decimal is implemented as a scaled int itself.  We can't use it though
// because the scale (Exponent) can potentially be anything.  This method normalizes the scale to the desired
// length of the fraction.
func decimalToScaledInt(d decimal.Decimal, minFracLen, maxFracLen int) *scaledInt64 {
	// round excess fraction part
	d = d.Round(int32(maxFracLen))

	co := d.Coefficient().Int64()
	ex := int(d.Exponent())
	frac := 0
	if d.Exponent() < 0 {
		// calculate the size of the fractional part
		frac = ex * -1
	} else if ex > 0 {
		// not sure if this can happen, but scale the coefficient to frac 0
		co *= int64(math.Pow10(ex))
	}

	// remove excess trailing zeros
	for co%10 == 0 && frac > minFracLen {
		co /= 10
		frac--
	}

	// add trailing zeros
	if frac < minFracLen {
		co *= int64(math.Pow10(minFracLen - frac))
		frac = minFracLen
	}

	return &scaledInt64{co, frac}
}

func getCurrencyFormatter(currency string) *money.Formatter {
	formatter := money.GetCurrency(currency).Formatter()
	if currency != "USD" && formatter.Grapheme == "$" {
		// This currency uses the $ symbol.  Append the currency code just to be clear it's not USD.
		formatter.Template += " " + currency
	}

	return formatter
}
