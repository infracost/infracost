package costs

import (
	"infracost/pkg/resource"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/shopspring/decimal"
)

func TestGenerateCostBreakdowns(t *testing.T) {
	r1 := resource.NewBaseResource("r1", map[string]interface{}{}, true)
	r1sr1 := resource.NewBaseResource("r1sr1", map[string]interface{}{}, true)
	r1.AddSubResource(r1sr1)
	r2 := resource.NewBaseResource("r2", map[string]interface{}{}, true)

	r1pc1 := resource.NewBasePriceComponent("r1pc1", r1, "r1pc1 unit", "hour", nil, nil)
	r1.AddPriceComponent(r1pc1)

	sr1pc1 := resource.NewBasePriceComponent("sr1pc1", r1sr1, "sr1pc1 unit", "hour", nil, nil)
	r1sr1.AddPriceComponent(sr1pc1)

	r2pc1 := resource.NewBasePriceComponent("r2pc1", r2, "r2pc1 unit", "hour", nil, nil)
	r2.AddPriceComponent(r2pc1)

	r2pc2 := resource.NewBasePriceComponent("r2pc2", r2, "r2pc2 unit", "hour", nil, nil)
	r2.AddPriceComponent(r2pc2)

	priceOverrides := make(map[resource.Resource]map[resource.PriceComponent]decimal.Decimal)
	priceOverrides[r1] = make(map[resource.PriceComponent]decimal.Decimal)
	priceOverrides[r1sr1] = make(map[resource.PriceComponent]decimal.Decimal)
	priceOverrides[r2] = make(map[resource.PriceComponent]decimal.Decimal)
	priceOverrides[r1][r1pc1] = decimal.NewFromFloat(float64(0.1))
	priceOverrides[r1sr1][sr1pc1] = decimal.NewFromFloat(float64(0.2))
	priceOverrides[r2][r2pc1] = decimal.NewFromFloat(float64(0.3))
	priceOverrides[r2][r2pc2] = decimal.NewFromFloat(float64(0.4))

	q := &testQueryRunner{
		priceOverrides: priceOverrides,
	}

	expected := []ResourceCostBreakdown{
		{
			Resource: r1,
			PriceComponentCosts: []PriceComponentCost{
				{
					PriceComponent: r1pc1,
					HourlyCost:     decimal.NewFromFloat(float64(0.1)),
					MonthlyCost:    decimal.NewFromFloat(float64(0.1)).Mul(decimal.NewFromInt(int64(730))),
				},
			},
			SubResourceCosts: []ResourceCostBreakdown{
				{
					Resource: r1sr1,
					PriceComponentCosts: []PriceComponentCost{
						{
							PriceComponent: sr1pc1,
							HourlyCost:     decimal.NewFromFloat(float64(0.2)),
							MonthlyCost:    decimal.NewFromFloat(float64(0.2)).Mul(decimal.NewFromInt(int64(730))),
						},
					},
					SubResourceCosts: []ResourceCostBreakdown{},
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

	result, err := GenerateCostBreakdowns(q, []resource.Resource{r1, r2})
	if err != nil {
		t.Error("received error", err)
	}

	if !cmp.Equal(result, expected, cmp.AllowUnexported(resource.BaseResource{}, resource.BasePriceComponent{})) {
		t.Error("got unexpected output", result)
	}
}
