package testutil

import (
	"fmt"
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/stretchr/testify/assert"

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
	return func(t *testing.T, c *schema.CostComponent) {
		assert.Equal(t, formatAmount(c.Price().Mul(multiplier)), formatAmount(c.HourlyCost()), fmt.Sprintf("unexpected hourly cost for %s", c.Name))
	}
}

func MonthlyPriceMultiplierCheck(multiplier decimal.Decimal) CostCheckFunc {
	return func(t *testing.T, c *schema.CostComponent) {
		assert.Equal(t, formatAmount(c.Price().Mul(multiplier)), formatAmount(c.MonthlyCost()), fmt.Sprintf("unexpected monthly cost for %s", c.Name))
	}
}

func TestResources(t *testing.T, resources []*schema.Resource, checks []ResourceCheck) {
	foundResources := make(map[*schema.Resource]bool)

	for _, check := range checks {
		found, r := findResource(resources, check.Name)
		assert.True(t, found, fmt.Sprintf("resource %s not found", check.Name))
		if !found {
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
		if r.NoPrice {
			continue
		}

		m, ok := foundResources[r]
		assert.True(t, ok && m, fmt.Sprintf("unexpected resource %s", r.Name))
	}
}

func TestCostComponents(t *testing.T, costComponents []*schema.CostComponent, checks []CostComponentCheck) {
	foundCostComponents := make(map[*schema.CostComponent]bool)

	for _, check := range checks {
		found, c := findCostComponent(costComponents, check.Name)
		assert.True(t, found, fmt.Sprintf("cost component %s not found", check.Name))
		if !found {
			continue
		}

		foundCostComponents[c] = true

		if check.SkipCheck {
			continue
		}

		assert.Equal(t, check.PriceHash, c.PriceHash(), fmt.Sprintf("unexpected price hash for %s", c.Name))

		if check.HourlyCostCheck != nil {
			check.HourlyCostCheck(t, c)
		}

		if check.MonthlyCostCheck != nil {
			check.MonthlyCostCheck(t, c)
		}
	}

	for _, c := range costComponents {
		m, ok := foundCostComponents[c]
		assert.True(t, ok && m, fmt.Sprintf("unexpected cost component %s", c.Name))
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

func formatAmount(d decimal.Decimal) string {
	f, _ := d.Float64()
	return fmt.Sprintf("%.4f", f)
}
