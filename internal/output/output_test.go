package output

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestCalculateTotalCosts(t *testing.T) {
	resources := []Resource{
		{
			HourlyCost:       decimalPtr(decimal.NewFromInt(10)),
			MonthlyCost:      decimalPtr(decimal.NewFromInt(7200)),
			MonthlyUsageCost: decimalPtr(decimal.NewFromInt(1000)),
		},
		{
			HourlyCost:       decimalPtr(decimal.NewFromInt(5)),
			MonthlyCost:      decimalPtr(decimal.NewFromInt(3600)),
			MonthlyUsageCost: decimalPtr(decimal.NewFromInt(500)),
		},
		{
			HourlyCost:       nil,
			MonthlyCost:      nil,
			MonthlyUsageCost: nil,
		},
	}

	totalHourlyCost, totalMonthlyCost, totalMonthlyUsageCost := calculateTotalCosts(resources)

	expected, _ := decimal.NewFromInt(15).Float64()
	actual, _ := totalHourlyCost.Float64()
	assert.Equal(t, expected, actual)
	expected, _ = decimal.NewFromInt(10800).Float64()
	actual, _ = totalMonthlyCost.Float64()
	assert.Equal(t, expected, actual)
	expected, _ = decimal.NewFromInt(1500).Float64()
	actual, _ = totalMonthlyUsageCost.Float64()
	assert.Equal(t, expected, actual)
}
