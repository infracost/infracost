package outputs

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
	Resource  string                   `json:"resource"`
	Breakdown []PriceComponentCostJSON `json:"breakdown"`
}

func ToJSON(resourceCostBreakdowns []base.ResourceCostBreakdown) ([]byte, error) {
	jsonObj := make([]ResourceCostBreakdownJSON, 0, len(resourceCostBreakdowns))

	for _, breakdown := range resourceCostBreakdowns {
		priceComponentCostJSONs := make([]PriceComponentCostJSON, 0, len(breakdown.PriceComponentCosts))
		for _, priceComponentCost := range breakdown.PriceComponentCosts {
			priceComponentCostJSONs = append(priceComponentCostJSONs, PriceComponentCostJSON{
				PriceComponent: priceComponentCost.PriceComponent.Name(),
				HourlyCost:     priceComponentCost.HourlyCost,
				MonthlyCost:    priceComponentCost.MonthlyCost,
			})
		}

		breakdownJSON := ResourceCostBreakdownJSON{
			Resource:  breakdown.Resource.Address(),
			Breakdown: priceComponentCostJSONs,
		}
		jsonObj = append(jsonObj, breakdownJSON)
	}

	return json.Marshal(jsonObj)
}
