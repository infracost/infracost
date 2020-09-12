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
	comps := make([]costComponentJSON, 0, len(resource.CostComponents))
	for _, c := range resource.CostComponents {
		comps = append(comps, costComponentJSON{
			Name:            c.Name,
			Unit:            c.Unit,
			HourlyQuantity:  c.HourlyQuantity.String(),
			MonthlyQuantity: c.MonthlyQuantity.String(),
			Price:           c.Price().String(),
			HourlyCost:      c.HourlyCost().String(),
			MonthlyCost:     c.MonthlyCost().String(),
		})
	}

	res := make([]resourceJSON, 0, len(resource.SubResources))
	for _, r := range resource.SubResources {
		res = append(res, newResourceJSON(r))
	}

	return resourceJSON{
		Name:           resource.Name,
		HourlyCost:     resource.HourlyCost().String(),
		MonthlyCost:    resource.MonthlyCost().String(),
		CostComponents: comps,
		SubResources:   res,
	}
}

func ToJSON(r []*schema.Resource) ([]byte, error) {
	s := make([]resourceJSON, 0, len(r))

	for _, res := range r {
		s = append(s, newResourceJSON(res))
	}

	return json.Marshal(s)
}
