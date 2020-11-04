package output

import (
	"encoding/json"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/urfave/cli/v2"
)

type outputJSON struct {
	Resources *[]resourceJSON `json:"resources"`
	Warnings  []string        `json:"warnings"`
}
type costComponentJSON struct {
	Name            string  `json:"name"`
	Unit            string  `json:"unit"`
	HourlyQuantity  *string `json:"hourlyQuantity"`
	MonthlyQuantity *string `json:"monthlyQuantity"`
	Price           string  `json:"price"`
	HourlyCost      *string `json:"hourlyCost"`
	MonthlyCost     *string `json:"monthlyCost"`
}

type resourceJSON struct {
	Name           string              `json:"name"`
	HourlyCost     *string             `json:"hourlyCost"`
	MonthlyCost    *string             `json:"monthlyCost"`
	CostComponents []costComponentJSON `json:"costComponents,omitempty"`
	SubResources   []resourceJSON      `json:"subresources,omitempty"`
}

func newResourceJSON(r *schema.Resource) resourceJSON {
	comps := make([]costComponentJSON, 0, len(r.CostComponents))
	for _, c := range r.CostComponents {

		comps = append(comps, costComponentJSON{
			Name:            c.Name,
			Unit:            c.UnitWithMultiplier(),
			HourlyQuantity:  strOrNil(c.UnitMultiplierHourlyQuantity()),
			MonthlyQuantity: strOrNil(c.UnitMultiplierMonthlyQuantity()),
			Price:           c.UnitMultiplierPrice().String(),
			HourlyCost:      strOrNil(c.HourlyCost),
			MonthlyCost:     strOrNil(c.MonthlyCost),
		})
	}

	subresources := make([]resourceJSON, 0, len(r.SubResources))
	for _, s := range r.SubResources {
		subresources = append(subresources, newResourceJSON(s))
	}

	return resourceJSON{
		Name:           r.Name,
		HourlyCost:     strOrNil(r.HourlyCost),
		MonthlyCost:    strOrNil(r.MonthlyCost),
		CostComponents: comps,
		SubResources:   subresources,
	}
}

func ToJSON(resources []*schema.Resource, c *cli.Context) ([]byte, error) {
	arr := make([]resourceJSON, 0, len(resources))

	for _, r := range resources {
		if r.IsSkipped {
			continue
		}
		arr = append(arr, newResourceJSON(r))
	}

	out := outputJSON{
		Resources: &arr,
	}

	msg := skippedResourcesMessage(resources, c.Bool("show-skipped"))
	if msg != "" {
		out.Warnings = append(out.Warnings, msg)
	}

	return json.Marshal(out)
}

func strOrNil(d *decimal.Decimal) *string {
	if d == nil {
		return nil
	}
	s := d.String()
	return &s
}
