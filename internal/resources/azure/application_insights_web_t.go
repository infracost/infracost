package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type ApplicationInsightsWebTest struct {
	Address string
	Region  string
	Kind    string
	Enabled bool
}

func (r *ApplicationInsightsWebTest) CoreType() string {
	return "ApplicationInsightsWebTest"
}

func (r *ApplicationInsightsWebTest) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *ApplicationInsightsWebTest) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ApplicationInsightsWebTest) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{}

	if r.Kind != "" {
		if strings.ToLower(r.Kind) == "multistep" && r.Enabled {
			costComponents = append(costComponents, &schema.CostComponent{
				Name:            "Multi-step web test",
				Unit:            "test",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("azure"),
					Region:        strPtr(r.Region),
					Service:       strPtr("Application Insights"),
					ProductFamily: strPtr("Management and Governance"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "meterName", ValueRegex: strPtr(fmt.Sprintf("/^%s$/i", "Multi-step Web Test"))},
						{Key: "skuName", ValueRegex: strPtr(fmt.Sprintf("/^%s$/i", "Enterprise"))},
					},
				},
			})
		}
	}

	if len(costComponents) == 0 {
		return &schema.Resource{
			Name:        r.Address,
			IsSkipped:   true,
			NoPrice:     true,
			UsageSchema: r.UsageSchema(),
		}
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}

}
