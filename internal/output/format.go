package output

import (
	"github.com/dustin/go-humanize"
	"github.com/shopspring/decimal"
)

var roundCostsAbove = 100

func formatQuantity(q *decimal.Decimal) string {
	if q == nil {
		return "-"
	}
	f, _ := q.Float64()
	return humanize.CommafWithDigits(f, 4)
}

func formatCost(d *decimal.Decimal) string {
	if d == nil {
		return "-"
	}

	f, _ := d.Float64()

	s := humanize.FormatFloat("#,###.##", f)
	if d.GreaterThanOrEqual(decimal.NewFromInt(int64(roundCostsAbove))) {
		s = humanize.FormatFloat("#,###.", f)
	}

	return "$" + s
}

func formatCost2DP(d *decimal.Decimal) string {
	if d == nil {
		return "-"
	}

	f, _ := d.Float64()

	s := humanize.FormatFloat("#,###.##", f)
	return "$" + s
}

func formatPrice(d decimal.Decimal) string {
	if d.LessThan(decimal.NewFromFloat(0.01)) {
		return "$" + d.String()
	}

	f, _ := d.Float64()

	s := humanize.FormatFloat("#,###.##", f)
	return "$" + s
}
