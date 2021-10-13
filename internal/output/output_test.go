package output

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestCalculateTotalCosts(t *testing.T) {
	resources := []Resource{
		{
			HourlyCost:  decimalPtr(decimal.NewFromInt(10)),
			MonthlyCost: decimalPtr(decimal.NewFromInt(7200)),
		},
		{
			HourlyCost:  decimalPtr(decimal.NewFromInt(5)),
			MonthlyCost: decimalPtr(decimal.NewFromInt(3600)),
		},
		{
			HourlyCost:  nil,
			MonthlyCost: nil,
		},
	}

	totalHourlyCost, totalMonthlyCost := calculateTotalCosts(resources)

	expected, _ := decimal.NewFromInt(15).Float64()
	actual, _ := totalHourlyCost.Float64()
	assert.Equal(t, expected, actual)
	expected, _ = decimal.NewFromInt(10800).Float64()
	actual, _ = totalMonthlyCost.Float64()
	assert.Equal(t, expected, actual)
}
