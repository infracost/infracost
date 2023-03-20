package azure

import (
	"fmt"
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// ApplicationInsightsStandardWebTest struct represents an Application Insights Standard WebTest.
//
// Resource information: https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/application_insights_standard_web_test
// Pricing information: https://azure.microsoft.com/en-in/pricing/details/monitor/
type ApplicationInsightsStandardWebTest struct {
	Address string
	Region  string

	Enabled   bool
	Frequency int64
}

// CoreType returns the name of this resource type
func (r *ApplicationInsightsStandardWebTest) CoreType() string {
	return "ApplicationInsightsStandardWebTest"
}

// UsageSchema defines a list which represents the usage schema of ApplicationInsightsStandardWebTest.
func (r *ApplicationInsightsStandardWebTest) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

// PopulateUsage parses the u schema.UsageData into the ApplicationInsightsStandardWebTest.
// It uses the `infracost_usage` struct tags to populate data into the ApplicationInsightsStandardWebTest.
func (r *ApplicationInsightsStandardWebTest) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid ApplicationInsightsStandardWebTest struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *ApplicationInsightsStandardWebTest) BuildResource() *schema.Resource {
	var costComponents []*schema.CostComponent

	if r.Enabled {
		secondsPerMonth := int64(730 * 60 * 60) // 730 hours * 60 minutes * 60 seconds
		tests := secondsPerMonth / r.Frequency

		costComponents = append(costComponents, &schema.CostComponent{
			Name:            fmt.Sprintf("Standard web test (%d second frequency)", r.Frequency),
			Unit:            "tests",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(tests)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("azure"),
				Region:        strPtr(r.Region),
				Service:       strPtr("Azure Monitor"),
				ProductFamily: strPtr("Management and Governance"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "skuName", Value: strPtr("Standard Web Test")},
				},
			},
		})
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}
