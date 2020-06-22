package base

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/shopspring/decimal"
)

func TestGenerateCostBreakdowns(t *testing.T) {
	sr1pc1 := &testPriceComponent{hourlyCost: decimal.NewFromFloat(float64(0.1))}
	r1pc1 := &testPriceComponent{hourlyCost: decimal.NewFromFloat(float64(0.2))}
	r2pc1 := &testPriceComponent{hourlyCost: decimal.NewFromFloat(float64(0.3))}
	r2pc2 := &testPriceComponent{hourlyCost: decimal.NewFromFloat(float64(0.4))}

	r1sr1 := &testResource{
		priceComponents: []PriceComponent{sr1pc1},
	}
	r1 := &testResource{
		priceComponents: []PriceComponent{r1pc1},
		subResources:    []Resource{r1sr1},
	}
	r2 := &testResource{
		priceComponents: []PriceComponent{r2pc1, r2pc2},
	}

	expected := []ResourceCostBreakdown{
		{
			Resource: r1,
			PriceComponentCosts: []PriceComponentCost{
				{
					PriceComponent: r1pc1,
					HourlyCost:     decimal.NewFromFloat(float64(0.2)),
					MonthlyCost:    decimal.NewFromFloat(float64(0.2)).Mul(decimal.NewFromInt(int64(730))),
				},
			},
			SubResourceCosts: []ResourceCostBreakdown{
				{
					Resource: r1sr1,
					PriceComponentCosts: []PriceComponentCost{
						{
							PriceComponent: sr1pc1,
							HourlyCost:     decimal.NewFromFloat(float64(0.1)),
							MonthlyCost:    decimal.NewFromFloat(float64(0.1)).Mul(decimal.NewFromInt(int64(730))),
						},
					},
				},
			},
		},
		{
			Resource: r2,
			PriceComponentCosts: []PriceComponentCost{
				{
					PriceComponent: r2pc1,
					HourlyCost:     decimal.NewFromFloat(float64(0.3)),
					MonthlyCost:    decimal.NewFromFloat(float64(0.3)).Mul(decimal.NewFromInt(int64(730))),
				},
				{
					PriceComponent: r2pc2,
					HourlyCost:     decimal.NewFromFloat(float64(0.4)),
					MonthlyCost:    decimal.NewFromFloat(float64(0.4)).Mul(decimal.NewFromInt(int64(730))),
				},
			},
			SubResourceCosts: []ResourceCostBreakdown{},
		},
	}

	priceOverrides := make(map[Resource]map[PriceComponent]decimal.Decimal)
	priceOverrides[r1] = make(map[PriceComponent]decimal.Decimal)
	priceOverrides[r1][r1pc1] = decimal.NewFromFloat(float64(0.01))

	q := &testQueryRunner{
		priceOverrides: priceOverrides,
	}
	result, err := GenerateCostBreakdowns(q, []Resource{r1, r2})
	if err != nil {
		t.Error("received error", err)
	}

	if !cmp.Equal(result, expected, cmp.AllowUnexported(testResource{}, testPriceComponent{})) {
		t.Error("got unexpected output", result)
	}

	if !cmp.Equal(r1pc1.price, decimal.NewFromFloat(float64(0.01))) {
		t.Error("did not set the correct price", r1pc1.price)
	}
}
