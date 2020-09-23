package output

import (
	"encoding/json"
	"fmt"

	"github.com/infracost/infracost/internal/providers/terraform"
	"github.com/infracost/infracost/pkg/schema"
	"github.com/urfave/cli/v2"
)

type outputJSON struct {
	Resources *[]resourceJSON `json:"resources"`
	Warnings  []string        `json:"warnings"`
}
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

func newResourceJSON(r *schema.Resource) resourceJSON {
	comps := make([]costComponentJSON, 0, len(r.CostComponents))
	for _, c := range r.CostComponents {
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

	subresources := make([]resourceJSON, 0, len(r.SubResources))
	for _, s := range r.SubResources {
		subresources = append(subresources, newResourceJSON(s))
	}

	return resourceJSON{
		Name:           r.Name,
		HourlyCost:     r.HourlyCost().String(),
		MonthlyCost:    r.MonthlyCost().String(),
		CostComponents: comps,
		SubResources:   subresources,
	}
}

func showSkippedResourcesJSON(resources []*schema.Resource, showDetails bool) []string {
	unSupportedTypeCount, _, unSupportedCount, _ := terraform.CountSkippedResources(resources)
	if unSupportedCount == 0 {
		return nil
	}
	message := fmt.Sprintf("\n%d out of %d resources couldn't be estimated as Infracost doesn't support them yet (https://www.infracost.io/docs/supported_resources)", unSupportedCount, len(resources))
	if showDetails {
		message += ".\n"
	} else {
		message += ", re-run with --show-skipped to see the list.\n"
	}
	message += "We're continually adding new resources, please create an issue if you'd like us to prioritize your list.\n"
	if showDetails {
		for rType, count := range unSupportedTypeCount {
			message += fmt.Sprintf("%d x %s\n", count, rType)
		}
	}
	return []string{message}
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

	out.Warnings = append(out.Warnings, showSkippedResourcesJSON(resources, c.Bool("show-skipped"))...)

	return json.Marshal(out)
}
