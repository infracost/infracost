package testutil

import (
	"fmt"
	"infracost/pkg/schema"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/shopspring/decimal"
)

var PriceHashResultSort = cmpopts.SortSlices(func(x, y []string) bool {
	return fmt.Sprintf("%s %s", x[0], x[1]) < fmt.Sprintf("%s %s", y[0], y[1])
})

func CheckPriceHashes(t *testing.T, resources []*schema.Resource, expectedPriceHashes [][]string) {
	priceHashResults := ExtractPriceHashes(resources)
	if !cmp.Equal(priceHashResults, expectedPriceHashes, PriceHashResultSort) {
		t.Error("Got unexpected price hashes", priceHashResults)
	}
}

func CheckCost(t *testing.T, resourceName string, costComponent *schema.CostComponent, costTimeUnit string, expectedCost decimal.Decimal) {
	var actualCost decimal.Decimal
	if costTimeUnit == "hourly" {
		actualCost = costComponent.HourlyCost()
	} else if costTimeUnit == "monthly" {
		actualCost = costComponent.MonthlyCost()
	} else {
		t.Error("Got unexpected costTimeUnit, expecting 'hourly' or 'monthly'")
	}

	if !cmp.Equal(actualCost, expectedCost) {
		t.Errorf("Got unexpected cost for %s -> %s: %s", resourceName, costComponent.Name, formatDecimal(actualCost, "%.4f"))
	}
}

func ExtractPriceHashes(resources []*schema.Resource) [][]string {
	priceHashes := [][]string{}

	for _, resource := range resources {
		for _, costComponent := range resource.CostComponents {
			priceHashes = append(priceHashes, []string{resource.Name, costComponent.Name, costComponent.PriceHash()})
		}

		priceHashes = append(priceHashes, ExtractPriceHashes(resource.SubResources)...)
	}

	return priceHashes
}

func FindCostComponent(resources []*schema.Resource, resourceName string, costComponentName string) *schema.CostComponent {
	for _, resource := range resources {
		for _, costComponent := range resource.CostComponents {
			if resource.Name == resourceName && costComponent.Name == costComponentName {
				return costComponent
			}
		}

		costComponent := FindCostComponent(resource.SubResources, resourceName, costComponentName)
		if costComponent != nil {
			return costComponent
		}
	}

	return nil
}

func formatDecimal(d decimal.Decimal, format string) string {
	f, _ := d.Float64()
	return fmt.Sprintf(format, f)
}
