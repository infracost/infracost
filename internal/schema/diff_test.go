package schema

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestCalculateDiff(t *testing.T) {
	pastResources := []*Resource{
		{
			Name:             "rs1",
			HourlyCost:       decimalPtr(decimal.NewFromInt(5)),
			MonthlyCost:      decimalPtr(decimal.NewFromInt(3600)),
			MonthlyUsageCost: decimalPtr(decimal.NewFromInt(600)),
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
	}

	currentResources := []*Resource{
		{
			Name:             "rs1",
			HourlyCost:       decimalPtr(decimal.NewFromInt(2)),
			MonthlyCost:      decimalPtr(decimal.NewFromInt(1440)),
			MonthlyUsageCost: decimalPtr(decimal.NewFromInt(200)),
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
	}

	expectedDiff := []*Resource{
		{
			Name:             "rs1",
			HourlyCost:       decimalPtr(decimal.NewFromInt(-3)),
			MonthlyCost:      decimalPtr(decimal.NewFromInt(-2160)),
			MonthlyUsageCost: decimalPtr(decimal.NewFromInt(-400)),
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
			Name:             "rs2",
			HourlyCost:       decimalPtr(decimal.NewFromInt(-1)),
			MonthlyCost:      decimalPtr(decimal.NewFromInt(-720)),
			MonthlyUsageCost: decimalPtr(decimal.Zero),
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
			Name:             "rs3",
			HourlyCost:       decimalPtr(decimal.NewFromInt(3)),
			MonthlyCost:      decimalPtr(decimal.NewFromInt(2160)),
			MonthlyUsageCost: decimalPtr(decimal.Zero),
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
	}

	diff := CalculateDiff(pastResources, currentResources)
	assert.Equal(t, expectedDiff, diff)
}

func TestDiffCostComponentsByResource(t *testing.T) {
	pastRS := &Resource{
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
			{
				Name:                "cc4 (label 1, label 2)",
				HourlyQuantity:      decimalPtr(decimal.NewFromInt(10)),
				MonthlyQuantity:     decimalPtr(decimal.NewFromInt(7200)),
				MonthlyDiscountPerc: 0.2,
				price:               decimal.NewFromInt(2),
				HourlyCost:          decimalPtr(decimal.NewFromInt(5)),
				MonthlyCost:         decimalPtr(decimal.NewFromInt(3600)),
			},
		},
	}
	currentRS := &Resource{
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
			{
				Name:                "cc4 (label 1, label 3)",
				HourlyQuantity:      decimalPtr(decimal.NewFromInt(20)),
				MonthlyQuantity:     decimalPtr(decimal.NewFromInt(14400)),
				MonthlyDiscountPerc: 0.45,
				price:               decimal.NewFromInt(3),
				HourlyCost:          decimalPtr(decimal.NewFromInt(2)),
				MonthlyCost:         decimalPtr(decimal.NewFromInt(1440)),
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
			Name:                "cc4 (label 1, label 2 → label 3)",
			HourlyQuantity:      decimalPtr(decimal.NewFromInt(10)),
			MonthlyQuantity:     decimalPtr(decimal.NewFromInt(7200)),
			MonthlyDiscountPerc: 0.25,
			price:               decimal.NewFromInt(1),
			HourlyCost:          decimalPtr(decimal.NewFromInt(-3)),
			MonthlyCost:         decimalPtr(decimal.NewFromInt(-2160)),
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

	changed, diff := diffCostComponentsByResource(pastRS, currentRS)
	assert.Equal(t, true, changed)
	assert.Equal(t, expectedDiff, diff)
}

func TestDiffDecimals(t *testing.T) {
	dc1 := decimalPtr(decimal.NewFromInt(10))
	dc2 := decimalPtr(decimal.NewFromInt(20))

	assert.Equal(t, decimal.Zero, *diffDecimals(dc1, dc1))
	assert.Equal(t, decimal.NewFromInt(10), *diffDecimals(dc2, dc1))
	assert.Equal(t, decimal.NewFromInt(-10), *diffDecimals(dc1, dc2))
	assert.Equal(t, decimal.Zero, *diffDecimals(nil, nil))
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

func TestFindMatchingCostComponent(t *testing.T) {
	tests := []struct {
		costComponents []*CostComponent
		name           string
		expected       *CostComponent
	}{
		{
			costComponents: []*CostComponent{
				{Name: "Instance usage"},
				{Name: "EBS-optimized usage"},
				{Name: "CPU Credits"},
			},
			name: "EBS-optimized usage",
			expected: &CostComponent{
				Name: "EBS-optimized usage",
			},
		},
		{
			costComponents: []*CostComponent{
				{Name: "Instance usage"},
				{Name: "EBS-optimized usage"},
				{Name: "CPU Credits"},
			},
			name:     "Not existent",
			expected: nil,
		},
		{
			costComponents: []*CostComponent{
				{Name: "Capacity (first 50TB)"},
				{Name: "Capacity (next 450TB)"},
				{Name: "Capacity (over 500TB)"},
			},
			name: "Capacity (over 500TB)",
			expected: &CostComponent{
				Name: "Capacity (over 500TB)",
			},
		},
		{
			costComponents: []*CostComponent{
				{Name: "Instance usage (Linux/UNIX, on-demand, t3.small)"},
				{Name: "EBS-optimized usage"},
				{Name: "CPU Credits"},
			},
			name: "Instance usage (Linux/UNIX, reserved, t3.small)",
			expected: &CostComponent{
				Name: "Instance usage (Linux/UNIX, on-demand, t3.small)",
			},
		},
		{
			costComponents: []*CostComponent{
				{Name: "Instance usage"},
				{Name: "EBS-optimized usage"},
				{Name: "CPU Credits"},
			},
			name:     "Instance usage (Linux/UNIX, reserved, t3.small)",
			expected: nil,
		},
	}

	for _, test := range tests {
		actual := findMatchingCostComponent(test.costComponents, test.name)
		assert.Equal(t, test.expected, actual)
	}
}

func TestDiffName(t *testing.T) {
	tests := []struct {
		current  string
		past     string
		expected string
	}{
		{
			current:  "Instance usage",
			past:     "Instance usage",
			expected: "Instance usage",
		},
		{
			current:  "Instance usage",
			past:     "",
			expected: "Instance usage",
		},
		{
			current:  "",
			past:     "Instance usage",
			expected: "Instance usage",
		},
		{
			current:  "Instance usage",
			past:     "Different",
			expected: "Instance usage",
		},
		{
			current:  "Instance usage (Linux/UNIX)",
			past:     "Instance usage (Linux/UNIX)",
			expected: "Instance usage (Linux/UNIX)",
		},
		{
			current:  "Instance usage (Linux/UNIX, reserved)",
			past:     "Instance usage (Linux/UNIX, on-demand)",
			expected: "Instance usage (Linux/UNIX, on-demand → reserved)",
		},
		{
			current:  "Instance usage (Linux/UNIX, reserved, t3.medium)",
			past:     "Instance usage (Linux/UNIX, on-demand, t3.small)",
			expected: "Instance usage (Linux/UNIX, on-demand → reserved, t3.small → t3.medium)",
		},
		{
			current:  "Instance usage (Linux/UNIX, reserved, t3.medium)",
			past:     "Instance usage (Linux/UNIX, on-demand)",
			expected: "Instance usage (Linux/UNIX, on-demand) → (Linux/UNIX, reserved, t3.medium)",
		},
		{
			current:  "Instance usage (Linux/UNIX, reserved)",
			past:     "Instance usage (Linux/UNIX, on-demand, t3.small)",
			expected: "Instance usage (Linux/UNIX, on-demand, t3.small) → (Linux/UNIX, reserved)",
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, diffName(test.current, test.past))
	}
}
