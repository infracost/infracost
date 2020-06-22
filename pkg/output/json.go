package output

import (
	"encoding/json"
	"plancosts/pkg/base"

	"github.com/shopspring/decimal"
)

type PriceComponentCostJSON struct {
	PriceComponent string          `json:"priceComponent"`
	HourlyCost     decimal.Decimal `json:"hourlyCost"`
	MonthlyCost    decimal.Decimal `json:"monthlyCost"`
}

type ResourceCostBreakdownJSON struct {
	Resource     string                      `json:"resource"`
	Breakdown    []PriceComponentCostJSON    `json:"breakdown"`
	SubResources []ResourceCostBreakdownJSON `json:"subresources,omitempty"`
}

func createJSONObj(breakdown base.ResourceCostBreakdown) ResourceCostBreakdownJSON {
	priceComponentCostJSONs := make([]PriceComponentCostJSON, 0, len(breakdown.PriceComponentCosts))
	for _, priceComponentCost := range breakdown.PriceComponentCosts {
		priceComponentCostJSONs = append(priceComponentCostJSONs, PriceComponentCostJSON{
			PriceComponent: priceComponentCost.PriceComponent.Name(),
			HourlyCost:     priceComponentCost.HourlyCost.Round(6),
			MonthlyCost:    priceComponentCost.MonthlyCost.Round(6),
		})
	}

	subResourcesCostBreakdownJSONs := make([]ResourceCostBreakdownJSON, 0, len(breakdown.SubResourceCosts))
	for _, subResourceBreakdown := range breakdown.SubResourceCosts {
		subResourcesCostBreakdownJSONs = append(subResourcesCostBreakdownJSONs, createJSONObj(subResourceBreakdown))
	}

	return ResourceCostBreakdownJSON{
		Resource:     breakdown.Resource.Address(),
		Breakdown:    priceComponentCostJSONs,
		SubResources: subResourcesCostBreakdownJSONs,
	}
}

func ToJSON(resourceCostBreakdowns []base.ResourceCostBreakdown) ([]byte, error) {
	jsonObjs := make([]ResourceCostBreakdownJSON, 0, len(resourceCostBreakdowns))
	for _, breakdown := range resourceCostBreakdowns {
		jsonObjs = append(jsonObjs, createJSONObj(breakdown))
	}
	return json.Marshal(jsonObjs)
}
