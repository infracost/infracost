package azure

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"strings"

	"github.com/shopspring/decimal"
)

type ApplicationInsightsWebTest struct {
	Address string
	Region  string
	Kind    string
	Enabled bool
}

var ApplicationInsightsWebTestUsageSchema = []*schema.UsageItem{}

func (r *ApplicationInsightsWebTest) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ApplicationInsightsWebTest) BuildResource() *schema.Resource {
	region := r.Region
	costComponents := []*schema.CostComponent{}

	if r.Kind != "" {
		if strings.ToLower(r.Kind) == "multistep" && r.Enabled {
			costComponents = append(costComponents, appInsightCostComponents(
				region,
				"Multi-step web test",
				"test",
				"Multi-step Web Test",
				"Enterprise",
				decimalPtr(decimal.NewFromInt(1))))
		}
	}

	if len(costComponents) == 0 {
		return &schema.Resource{
			Name:      r.Address,
			IsSkipped: true,
			NoPrice:   true, UsageSchema: ApplicationInsightsWebTestUsageSchema,
		}
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents, UsageSchema: ApplicationInsightsWebTestUsageSchema,
	}

}
