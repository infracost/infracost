package testutil

import (
	"fmt"
	"infracost/pkg/schema"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/shopspring/decimal"
)

type CostCheckFunc func(*testing.T, *schema.CostComponent)

type ResourceCheck struct {
	Name                string
	CostComponentChecks []CostComponentCheck
	SubResourceChecks   []ResourceCheck
}

type CostComponentCheck struct {
	Name             string
	PriceHash        string
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

func TestResource(t *testing.T, resources []*schema.Resource, resourceCheck ResourceCheck) {
	found, resource := findResource(resources, resourceCheck.Name)
	if !found {
		t.Errorf("No resource matched for name %s", resourceCheck.Name)
		return
	}

	for _, costComponentCheck := range resourceCheck.CostComponentChecks {
		TestCostComponent(t, resource.CostComponents, costComponentCheck)
	}

	for _, subResourceCheck := range resourceCheck.SubResourceChecks {
		TestResource(t, resource.SubResources, subResourceCheck)
	}
}

func TestCostComponent(t *testing.T, costComponents []*schema.CostComponent, costComponentCheck CostComponentCheck) {
	found, costComponent := findCostComponent(costComponents, costComponentCheck.Name)
	if !found {
		t.Errorf("No cost componenet matched for name %s", costComponentCheck.Name)
		return
	}

	if !cmp.Equal(costComponent.PriceHash(), costComponentCheck.PriceHash) {
		t.Errorf("Unexpected cost component price hash for %s (expected: %s, got: %s)", costComponent.Name, costComponentCheck.PriceHash, costComponent.PriceHash())
	}

	if costComponentCheck.HourlyCostCheck != nil {
		costComponentCheck.HourlyCostCheck(t, costComponent)
	}

	if costComponentCheck.MonthlyCostCheck != nil {
		costComponentCheck.MonthlyCostCheck(t, costComponent)
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
