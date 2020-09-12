package output

import (
	"encoding/json"

	"github.com/infracost/infracost/pkg/schema"
)

type costComponentJSON struct {
	Name            string `json:"name"`
	Unit            string `json:"unit"`
	HourlyQuantity  string `json:"hourlyQuantity"`
	MonthlyQuantity string `json:"monthlyQuantity"`
	Price           string `json:"price"`
	HourlyCost      string `json:"hourlyCost"`
	MonthlyCost     string `json:"monthlyCost"`
}

type resourceJSON struct {
	Name           string              `json:"name"`
	HourlyCost     string              `json:"hourlyCost"`
	MonthlyCost    string              `json:"monthlyCost"`
	CostComponents []costComponentJSON `json:"costComponents,omitempty"`
	SubResources   []resourceJSON      `json:"subresources,omitempty"`
}

func newResourceJSON(resource *schema.Resource) resourceJSON {
	costComponentJSONs := make([]costComponentJSON, 0, len(resource.CostComponents))
	for _, costComponent := range resource.CostComponents {
		costComponentJSONs = append(costComponentJSONs, costComponentJSON{
			Name:            costComponent.Name,
			Unit:            costComponent.Unit,
			HourlyQuantity:  costComponent.HourlyQuantity.String(),
			MonthlyQuantity: costComponent.MonthlyQuantity.String(),
			Price:           costComponent.Price().String(),
			HourlyCost:      costComponent.HourlyCost().String(),
			MonthlyCost:     costComponent.MonthlyCost().String(),
		})
	}

	subResourceJSONs := make([]resourceJSON, 0, len(resource.SubResources))
	for _, subResource := range resource.SubResources {
		subResourceJSONs = append(subResourceJSONs, newResourceJSON(subResource))
	}

	return resourceJSON{
		Name:           resource.Name,
		HourlyCost:     resource.HourlyCost().String(),
		MonthlyCost:    resource.MonthlyCost().String(),
		CostComponents: costComponentJSONs,
		SubResources:   subResourceJSONs,
	}
}

func ToJSON(r []*schema.Resource) ([]byte, error) {
	s := make([]resourceJSON, 0, len(r))

	for _, res := range r {
		s = append(s, newResourceJSON(res))
	}

	return json.Marshal(s)
}
