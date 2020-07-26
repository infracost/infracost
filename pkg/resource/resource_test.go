package resource

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/shopspring/decimal"
)

func TestFlattenSubResources(t *testing.T) {
	r1 := NewBaseResource("r1", map[string]interface{}{}, true)
	r2 := NewBaseResource("r2", map[string]interface{}{}, true)
	r3 := NewBaseResource("r3", map[string]interface{}{}, true)
	r4 := NewBaseResource("r4", map[string]interface{}{}, true)
	r5 := NewBaseResource("r5", map[string]interface{}{}, true)
	r6 := NewBaseResource("r6", map[string]interface{}{}, true)

	r1.AddSubResource(r2)
	r1.AddSubResource(r3)
	r2.AddSubResource(r4)
	r2.AddSubResource(r5)
	r4.AddSubResource(r6)

	result := FlattenSubResources(r1)
	expected := []Resource{r2, r4, r6, r5, r3}
	if !cmp.Equal(result, expected, cmp.AllowUnexported(BaseResource{})) {
		t.Error("did not flatten subresources correctly", result)
	}
}

func TestBasePriceComponentQuantity(t *testing.T) {
	r1 := NewBaseResource("r1", map[string]interface{}{}, true)
	monthlyPc := NewBasePriceComponent("monthlyPc", r1, "unit", "month")
	hourlyPc := NewBasePriceComponent("hourlyPc", r1, "unit", "hour")

	result := monthlyPc.Quantity()
	if !cmp.Equal(result, decimal.NewFromInt(int64(1))) {
		t.Error("monthly quantity is incorrect", result)
	}

	result = hourlyPc.Quantity()
	if !cmp.Equal(result, decimal.NewFromInt(int64(730))) {
		t.Error("hourly quantity is incorrect", result)
	}

	r1.SetResourceCount(2)
	result = hourlyPc.Quantity()
	if !cmp.Equal(result, decimal.NewFromInt(int64(1460))) {
		t.Error("resource count does not change quantity correctly", result)
	}
	r1.SetResourceCount(1)

	hourlyPc.SetQuantityMultiplierFunc(func(resource Resource) decimal.Decimal {
		return decimal.NewFromInt(int64(3))
	})
	result = hourlyPc.Quantity()
	if !cmp.Equal(result, decimal.NewFromInt(int64(2190))) {
		t.Error("quantityMultiplierFunc count does not change quantity correctly", result)
	}
}

func TestBasePriceComponentHourlyCost(t *testing.T) {
	r1 := NewBaseResource("r1", map[string]interface{}{}, true)
	monthlyPc := NewBasePriceComponent("monthlyPc", r1, "unit", "month")
	hourlyPc := NewBasePriceComponent("hourlyPc", r1, "unit", "hour")

	result := hourlyPc.HourlyCost()
	if !cmp.Equal(result, decimal.Zero) {
		t.Error("cost should be 0 if no price is set", result)
	}

	hourlyPc.SetPrice(decimal.NewFromFloat(float64(0.2)))
	hourlyPc.SetQuantityMultiplierFunc(func(resource Resource) decimal.Decimal {
		return decimal.NewFromInt(int64(2))
	})
	result = hourlyPc.HourlyCost()
	if !cmp.Equal(result, decimal.NewFromFloat(float64(0.4))) {
		t.Error("hourly cost incorrect for hourly price component", result)
	}

	monthlyPc.SetPrice(decimal.NewFromFloat(float64(7.3)))
	monthlyPc.SetQuantityMultiplierFunc(func(resource Resource) decimal.Decimal {
		return decimal.NewFromInt(int64(4))
	})
	result = monthlyPc.HourlyCost()
	if !cmp.Equal(result, decimal.NewFromFloat(float64(0.04))) {
		t.Error("monthly cost incorrect for monthly price component", result)
	}
}

func TestBestResourceSubResources(t *testing.T) {
	r1 := NewBaseResource("r1", map[string]interface{}{}, true)
	r2 := NewBaseResource("charlie", map[string]interface{}{}, true)
	r3 := NewBaseResource("alpha", map[string]interface{}{}, true)
	r4 := NewBaseResource("bravo", map[string]interface{}{}, true)

	r1.AddSubResource(r2)
	r1.AddSubResource(r3)
	r1.AddSubResource(r4)

	result := r1.SubResources()
	expected := []Resource{r3, r4, r2}
	if !cmp.Equal(result, expected, cmp.AllowUnexported(BaseResource{})) {
		t.Error("did not sort the subresources correctly", result)
	}
}

func TestBestResourcePriceComponents(t *testing.T) {
	r1 := NewBaseResource("r1", map[string]interface{}{}, true)
	pc1 := NewBasePriceComponent("charlie", r1, "unit", "month")
	pc2 := NewBasePriceComponent("alpha", r1, "unit", "month")
	pc3 := NewBasePriceComponent("bravo", r1, "unit", "month")

	r1.AddPriceComponent(pc1)
	r1.AddPriceComponent(pc2)
	r1.AddPriceComponent(pc3)

	result := r1.PriceComponents()
	expected := []PriceComponent{pc2, pc3, pc1}
	if !cmp.Equal(result, expected, cmp.AllowUnexported(BaseResource{}, BasePriceComponent{}), cmpopts.IgnoreFields(BasePriceComponent{}, "resource")) {
		t.Error("did not sort the price component correctly", result)
	}
}
