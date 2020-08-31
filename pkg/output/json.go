package output

import (
	"encoding/json"
	"infracost/pkg/schema"

	"github.com/shopspring/decimal"
)

type costComponentJSON struct {
	Name            string          `json:"name"`
	Unit            string          `json:"unit"`
	HourlyQuantity  decimal.Decimal `json:"hourlyQuantity"`
	MonthlyQuantity decimal.Decimal `json:"monthlyQuantity"`
	Price           decimal.Decimal `json:"price"`
	HourlyCost      decimal.Decimal `json:"hourlyCost"`
	MonthlyCost     decimal.Decimal `json:"monthlyCost"`
}

type resourceJSON struct {
	Name           string              `json:"name"`
	HourlyCost     decimal.Decimal     `json:"hourlyCost"`
	MonthlyCost    decimal.Decimal     `json:"monthlyCost"`
	CostComponents []costComponentJSON `json:"costComponents,omitempty"`
	SubResources   []resourceJSON      `json:"subresources,omitempty"`
}

func newResourceJSON(resource *schema.Resource) resourceJSON {
	costComponentJSONs := make([]costComponentJSON, 0, len(resource.CostComponents))
	for _, costComponent := range resource.CostComponents {
		costComponentJSONs = append(costComponentJSONs, costComponentJSON{
			Name:            costComponent.Name,
			Unit:            costComponent.Unit,
			HourlyQuantity:  costComponent.HourlyQuantity.Round(6),
			MonthlyQuantity: costComponent.MonthlyQuantity.Round(6),
			Price:           costComponent.Price().Round(6),
			HourlyCost:      costComponent.HourlyCost().Round(6),
			MonthlyCost:     costComponent.MonthlyCost().Round(6),
		})
	}

	subResourceJSONs := make([]resourceJSON, 0, len(resource.SubResources))
	for _, subResource := range resource.SubResources {
		subResourceJSONs = append(subResourceJSONs, newResourceJSON(subResource))
	}

	return resourceJSON{
		Name:           resource.Name,
		HourlyCost:     resource.HourlyCost().Round(6),
		MonthlyCost:    resource.MonthlyCost().Round(6),
		CostComponents: costComponentJSONs,
		SubResources:   subResourceJSONs,
	}
}

func ToJSON(resources []*schema.Resource) ([]byte, error) {
	resourceJSONs := make([]resourceJSON, 0, len(resources))
	for _, resource := range resources {
		resourceJSONs = append(resourceJSONs, newResourceJSON(resource))
	}
	return json.Marshal(resourceJSONs)
}
