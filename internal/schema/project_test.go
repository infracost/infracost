package schema

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestProjectCalculateTotalCosts(t *testing.T) {
	existingResources := []*Resource{
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
	existingRs := &Breakdown{
		Resources: existingResources,
	}
	plannedResources := []*Resource{
		{
			HourlyCost:  decimalPtr(decimal.NewFromInt(20)),
			MonthlyCost: decimalPtr(decimal.NewFromInt(14400)),
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
	plannedRs := &Breakdown{
		Resources: plannedResources,
	}

	project := &Project{
		PastBreakdown: existingRs,
		Breakdown:     plannedRs,
	}
	project.CalculateTotalCosts()

	expected, _ := decimal.NewFromInt(15).Float64()
	actual, _ := project.PastBreakdown.TotalHourlyCost.Float64()
	assert.Equal(t, expected, actual)
	expected, _ = decimal.NewFromInt(10800).Float64()
	actual, _ = project.PastBreakdown.TotalMonthlyCost.Float64()
	assert.Equal(t, expected, actual)

	expected, _ = decimal.NewFromInt(25).Float64()
	actual, _ = project.Breakdown.TotalHourlyCost.Float64()
	assert.Equal(t, expected, actual)
	expected, _ = decimal.NewFromInt(18000).Float64()
	actual, _ = project.Breakdown.TotalMonthlyCost.Float64()
	assert.Equal(t, expected, actual)
}

func TestCalculateDiff(t *testing.T) {
	project := &Project{}
	project.PastBreakdown = &Breakdown{
		Resources: []*Resource{
			{
				Name:        "rs1",
				HourlyCost:  decimalPtr(decimal.NewFromInt(5)),
				MonthlyCost: decimalPtr(decimal.NewFromInt(3600)),
				CostComponents: []*CostComponent{
					{
						Name:                "cc1",
						HourlyQuantity:      decimalPtr(decimal.NewFromInt(10)),
						MonthlyQuantity:     decimalPtr(decimal.NewFromInt(7200)),
						MonthlyDiscountPerc: 0.2,
						price:               decimal.NewFromInt(2),
						HourlyCost:          decimalPtr(decimal.NewFromInt(5)),
						MonthlyCost:         decimalPtr(decimal.NewFromInt(3600)),
					},
				},
			},
			{
				Name:        "rs2",
				HourlyCost:  decimalPtr(decimal.NewFromInt(1)),
				MonthlyCost: decimalPtr(decimal.NewFromInt(720)),
				CostComponents: []*CostComponent{
					{
						Name:                "cc2",
						HourlyQuantity:      decimalPtr(decimal.NewFromInt(2)),
						MonthlyQuantity:     decimalPtr(decimal.NewFromInt(1440)),
						MonthlyDiscountPerc: 0.5,
						price:               decimal.NewFromInt(1),
						HourlyCost:          decimalPtr(decimal.NewFromInt(1)),
						MonthlyCost:         decimalPtr(decimal.NewFromInt(720)),
					},
				},
			},
		},
		TotalHourlyCost:  decimalPtr(decimal.NewFromInt(3)),
		TotalMonthlyCost: decimalPtr(decimal.NewFromInt(2160)),
	}
	project.Breakdown = &Breakdown{
		Resources: []*Resource{
			{
				Name:        "rs1",
				HourlyCost:  decimalPtr(decimal.NewFromInt(2)),
				MonthlyCost: decimalPtr(decimal.NewFromInt(1440)),
				CostComponents: []*CostComponent{
					{
						Name:                "cc1",
						HourlyQuantity:      decimalPtr(decimal.NewFromInt(20)),
						MonthlyQuantity:     decimalPtr(decimal.NewFromInt(14400)),
						MonthlyDiscountPerc: 0.45,
						price:               decimal.NewFromInt(3),
						HourlyCost:          decimalPtr(decimal.NewFromInt(2)),
						MonthlyCost:         decimalPtr(decimal.NewFromInt(1440)),
					},
				},
			},
			{
				Name:        "rs3",
				HourlyCost:  decimalPtr(decimal.NewFromInt(3)),
				MonthlyCost: decimalPtr(decimal.NewFromInt(2160)),
				CostComponents: []*CostComponent{
					{
						Name:                "cc3",
						HourlyQuantity:      decimalPtr(decimal.NewFromInt(3)),
						MonthlyQuantity:     decimalPtr(decimal.NewFromInt(2160)),
						MonthlyDiscountPerc: 0,
						price:               decimal.NewFromInt(3),
						HourlyCost:          decimalPtr(decimal.NewFromInt(3)),
						MonthlyCost:         decimalPtr(decimal.NewFromInt(2160)),
					},
				},
			},
		},
		TotalHourlyCost:  decimalPtr(decimal.NewFromInt(5)),
		TotalMonthlyCost: decimalPtr(decimal.NewFromInt(3600)),
	}

	expectedDiff := &Breakdown{
		Resources: []*Resource{
			{
				Name:        "rs1",
				HourlyCost:  decimalPtr(decimal.NewFromInt(-3)),
				MonthlyCost: decimalPtr(decimal.NewFromInt(-2160)),
				CostComponents: []*CostComponent{
					{
						Name:                "cc1",
						HourlyQuantity:      decimalPtr(decimal.NewFromInt(10)),
						MonthlyQuantity:     decimalPtr(decimal.NewFromInt(7200)),
						MonthlyDiscountPerc: 0.25,
						price:               decimal.NewFromInt(1),
						HourlyCost:          decimalPtr(decimal.NewFromInt(-3)),
						MonthlyCost:         decimalPtr(decimal.NewFromInt(-2160)),
					},
				},
			},
			{
				Name:        "rs2",
				HourlyCost:  decimalPtr(decimal.NewFromInt(-1)),
				MonthlyCost: decimalPtr(decimal.NewFromInt(-720)),
				CostComponents: []*CostComponent{
					{
						Name:                "cc2",
						HourlyQuantity:      decimalPtr(decimal.NewFromInt(-2)),
						MonthlyQuantity:     decimalPtr(decimal.NewFromInt(-1440)),
						MonthlyDiscountPerc: -0.5,
						price:               decimal.NewFromInt(-1),
						HourlyCost:          decimalPtr(decimal.NewFromInt(-1)),
						MonthlyCost:         decimalPtr(decimal.NewFromInt(-720)),
					},
				},
			},
			{
				Name:        "rs3",
				HourlyCost:  decimalPtr(decimal.NewFromInt(3)),
				MonthlyCost: decimalPtr(decimal.NewFromInt(2160)),
				CostComponents: []*CostComponent{
					{
						Name:                "cc3",
						HourlyQuantity:      decimalPtr(decimal.NewFromInt(3)),
						MonthlyQuantity:     decimalPtr(decimal.NewFromInt(2160)),
						MonthlyDiscountPerc: 0,
						price:               decimal.NewFromInt(3),
						HourlyCost:          decimalPtr(decimal.NewFromInt(3)),
						MonthlyCost:         decimalPtr(decimal.NewFromInt(2160)),
					},
				},
			},
		},
		TotalHourlyCost:  decimalPtr(decimal.NewFromInt(2)),
		TotalMonthlyCost: decimalPtr(decimal.NewFromInt(1440)),
	}

	project.CalculateDiff()
	assert.Equal(t, expectedDiff, project.Diff)
}

func TestDiffdiffCostComponentsByResource(t *testing.T) {
	existingRS := &Resource{
		Name: "rs",
		CostComponents: []*CostComponent{
			{
				Name:                "cc1",
				HourlyQuantity:      decimalPtr(decimal.NewFromInt(10)),
				MonthlyQuantity:     decimalPtr(decimal.NewFromInt(7200)),
				MonthlyDiscountPerc: 0.2,
				price:               decimal.NewFromInt(2),
				HourlyCost:          decimalPtr(decimal.NewFromInt(5)),
				MonthlyCost:         decimalPtr(decimal.NewFromInt(3600)),
			},
			{
				Name:                "cc2",
				HourlyQuantity:      decimalPtr(decimal.NewFromInt(2)),
				MonthlyQuantity:     decimalPtr(decimal.NewFromInt(1440)),
				MonthlyDiscountPerc: 0.5,
				price:               decimal.NewFromInt(1),
				HourlyCost:          decimalPtr(decimal.NewFromInt(1)),
				MonthlyCost:         decimalPtr(decimal.NewFromInt(720)),
			},
		},
	}
	plannedRS := &Resource{
		Name: "rs",
		CostComponents: []*CostComponent{
			{
				Name:                "cc1",
				HourlyQuantity:      decimalPtr(decimal.NewFromInt(20)),
				MonthlyQuantity:     decimalPtr(decimal.NewFromInt(14400)),
				MonthlyDiscountPerc: 0.45,
				price:               decimal.NewFromInt(3),
				HourlyCost:          decimalPtr(decimal.NewFromInt(2)),
				MonthlyCost:         decimalPtr(decimal.NewFromInt(1440)),
			},
			{
				Name:                "cc3",
				HourlyQuantity:      decimalPtr(decimal.NewFromInt(3)),
				MonthlyQuantity:     decimalPtr(decimal.NewFromInt(2160)),
				MonthlyDiscountPerc: 0,
				price:               decimal.NewFromInt(3),
				HourlyCost:          decimalPtr(decimal.NewFromInt(3)),
				MonthlyCost:         decimalPtr(decimal.NewFromInt(2160)),
			},
		},
	}

	expectedDiff := []*CostComponent{
		{
			Name:                "cc1",
			HourlyQuantity:      decimalPtr(decimal.NewFromInt(10)),
			MonthlyQuantity:     decimalPtr(decimal.NewFromInt(7200)),
			MonthlyDiscountPerc: 0.25,
			price:               decimal.NewFromInt(1),
			HourlyCost:          decimalPtr(decimal.NewFromInt(-3)),
			MonthlyCost:         decimalPtr(decimal.NewFromInt(-2160)),
		},
		{
			Name:                "cc2",
			HourlyQuantity:      decimalPtr(decimal.NewFromInt(-2)),
			MonthlyQuantity:     decimalPtr(decimal.NewFromInt(-1440)),
			MonthlyDiscountPerc: -0.5,
			price:               decimal.NewFromInt(-1),
			HourlyCost:          decimalPtr(decimal.NewFromInt(-1)),
			MonthlyCost:         decimalPtr(decimal.NewFromInt(-720)),
		},
		{
			Name:                "cc3",
			HourlyQuantity:      decimalPtr(decimal.NewFromInt(3)),
			MonthlyQuantity:     decimalPtr(decimal.NewFromInt(2160)),
			MonthlyDiscountPerc: 0,
			price:               decimal.NewFromInt(3),
			HourlyCost:          decimalPtr(decimal.NewFromInt(3)),
			MonthlyCost:         decimalPtr(decimal.NewFromInt(2160)),
		},
	}

	changed, diff := diffCostComponentsByResource(existingRS, plannedRS)
	assert.Equal(t, true, changed)
	assert.Equal(t, expectedDiff, diff)
}

func TestDiffDecimals(t *testing.T) {
	dc1 := decimalPtr(decimal.NewFromInt(10))
	dc2 := decimalPtr(decimal.NewFromInt(20))

	assert.Equal(t, decimal.Zero, *diffDecimals(dc1, dc1))
	assert.Equal(t, decimal.NewFromInt(10), *diffDecimals(dc2, dc1))
	assert.Equal(t, decimal.NewFromInt(-10), *diffDecimals(dc1, dc2))
}

func TestGetResourcesMap(t *testing.T) {
	rs1 := &Resource{
		Name: "rs1",
	}
	rs2 := &Resource{
		Name: "rs2",
		SubResources: []*Resource{
			{
				Name: "rs2_1",
				SubResources: []*Resource{
					{
						Name: "rs2_1_1",
					},
				},
			},
			{
				Name: "rs2_2",
			},
		},
	}
	resources := []*Resource{rs1, rs2}

	resourcesMap := make(map[string]*Resource)
	fillResourcesMap(resourcesMap, "", resources)
	expectedMap := map[string]*Resource{
		"rs1":               rs1,
		"rs2":               rs2,
		"rs2.rs2_1":         rs2.SubResources[0],
		"rs2.rs2_1.rs2_1_1": rs2.SubResources[0].SubResources[0],
		"rs2.rs2_2":         rs2.SubResources[1],
	}
	assert.Equal(t, expectedMap, resourcesMap)
}

func TestGetCostComponentsMap(t *testing.T) {
	cc1 := &CostComponent{
		Name: "cc1",
	}
	cc2 := &CostComponent{
		Name: "cc2",
	}
	resource := &Resource{
		CostComponents: []*CostComponent{cc1, cc2},
	}
	ccMap := getCostComponentsMap(resource)
	expectedMap := map[string]*CostComponent{
		"cc1": cc1,
		"cc2": cc2,
	}
	assert.Equal(t, expectedMap, ccMap)
}
