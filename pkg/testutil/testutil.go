package testutil

import (
	"fmt"
	"testing"

	"github.com/infracost/infracost/pkg/schema"

	"github.com/google/go-cmp/cmp"
	"github.com/shopspring/decimal"
)

type CostCheckFunc func(*testing.T, *schema.CostComponent)

type ResourceCheck struct {
	Name                string
	SkipCheck           bool
	CostComponentChecks []CostComponentCheck
	SubResourceChecks   []ResourceCheck
}

type CostComponentCheck struct {
	Name             string
	PriceHash        string
	SkipCheck        bool
	HourlyCostCheck  CostCheckFunc
	MonthlyCostCheck CostCheckFunc
}

func HourlyPriceMultiplierCheck(multiplier decimal.Decimal) CostCheckFunc {
	return func(t *testing.T, costComponent *schema.CostComponent) {
		expected := costComponent.Price().Mul(multiplier)
		if !cmp.Equal(costComponent.HourlyCost(), expected) {
			t.Errorf("Unexpected hourly cost for %s (expected: %s, got: %s)", costComponent.Name, formatDecimal(expected, "%.4f"), formatDecimal(costComponent.HourlyCost(), "%.4f"))
		}
	}
}

func MonthlyPriceMultiplierCheck(multiplier decimal.Decimal) CostCheckFunc {
	return func(t *testing.T, costComponent *schema.CostComponent) {
		expected := costComponent.Price().Mul(multiplier)
		if !cmp.Equal(costComponent.MonthlyCost(), expected) {
			t.Errorf("Unexpected monthly cost for %s (expected: %s, got: %s)", costComponent.Name, formatDecimal(expected, "%.4f"), formatDecimal(costComponent.MonthlyCost(), "%.4f"))
		}
	}
}

func TestResources(t *testing.T, resources []*schema.Resource, checks []ResourceCheck) {
	foundResources := make(map[*schema.Resource]bool)

	for _, check := range checks {
		found, r := findResource(resources, check.Name)
		if !found {
			t.Errorf("No resource matched for name %s", check.Name)
			continue
		}
		foundResources[r] = true

		if check.SkipCheck {
			continue
		}

		TestCostComponents(t, r.CostComponents, check.CostComponentChecks)
		TestResources(t, r.SubResources, check.SubResourceChecks)
	}

	for _, r := range resources {
		if m, ok := foundResources[r]; !ok || !m {
			t.Errorf("Unexpected resource %s", r.Name)
		}
	}
}

func TestCostComponents(t *testing.T, costComponents []*schema.CostComponent, checks []CostComponentCheck) {
	foundCostComponents := make(map[*schema.CostComponent]bool)

	for _, check := range checks {
		found, c := findCostComponent(costComponents, check.Name)
		if !found {
			t.Errorf("No cost component matched for name %s", check.Name)
			continue
		}
		foundCostComponents[c] = true

		if check.SkipCheck {
			continue
		}

		if !cmp.Equal(c.PriceHash(), check.PriceHash) {
			t.Errorf("Unexpected cost component price hash for %s (expected: %s, got: %s)", c.Name, check.PriceHash, c.PriceHash())
		}

		if check.HourlyCostCheck != nil {
			check.HourlyCostCheck(t, c)
		}

		if check.MonthlyCostCheck != nil {
			check.MonthlyCostCheck(t, c)
		}
	}

	for _, c := range costComponents {
		if m, ok := foundCostComponents[c]; !ok || !m {
			t.Errorf("Unexpected cost component %s", c.Name)
		}
	}
}

func findResource(resources []*schema.Resource, name string) (bool, *schema.Resource) {
	for _, resource := range resources {
		if resource.Name == name {
			return true, resource
		}
	}
	return false, nil
}

func findCostComponent(costComponents []*schema.CostComponent, name string) (bool, *schema.CostComponent) {
	for _, costComponent := range costComponents {
		if costComponent.Name == name {
			return true, costComponent
		}
	}
	return false, nil
}

func formatDecimal(d decimal.Decimal, format string) string {
	f, _ := d.Float64()
	return fmt.Sprintf(format, f)
}
