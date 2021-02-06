package schema

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestResourceStateCalculateTotalCosts(t *testing.T) {
	resources := []*Resource{
		{
			HourlyCost:  decimalPtr(decimal.NewFromInt(10)),
			MonthlyCost: decimalPtr(decimal.NewFromInt(7200)),
		},
		{
			HourlyCost:  decimalPtr(decimal.NewFromInt(5)),
			MonthlyCost: decimalPtr(decimal.NewFromInt(3600)),
		},
		{
			IsSkipped:   true,
			HourlyCost:  decimalPtr(decimal.NewFromInt(10)),
			MonthlyCost: decimalPtr(decimal.NewFromInt(7200)),
		},
		{
			HourlyCost:  nil,
			MonthlyCost: nil,
		},
	}
	rs := ResourcesState{
		Resources: resources,
	}
	rs.calculateTotalCosts()

	expected, _ := decimal.NewFromInt(15).Float64()
	actual, _ := rs.TotalHourlyCost.Float64()
	assert.Equal(t, expected, actual)

	expected, _ = decimal.NewFromInt(10800).Float64()
	actual, _ = rs.TotalMonthlyCost.Float64()
	assert.Equal(t, expected, actual)
}
