package azure

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"

	"github.com/shopspring/decimal"
)

type AutomationDSCConfiguration struct {
	Address                 string
	Region                  string
	NonAzureConfigNodeCount *int64 `infracost_usage:"non_azure_config_node_count"`
}

func (r *AutomationDSCConfiguration) CoreType() string {
	return "AutomationDSCConfiguration"
}

func (r *AutomationDSCConfiguration) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{{Key: "non_azure_config_node_count", ValueType: schema.Int64, DefaultValue: 0}}
}

func (r *AutomationDSCConfiguration) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *AutomationDSCConfiguration) BuildResource() *schema.Resource {
	return &schema.Resource{
		Name:           r.Address,
		CostComponents: automationDSCNodesCostComponent(&r.Region, r.NonAzureConfigNodeCount),
		UsageSchema:    r.UsageSchema(),
	}
}

func automationDSCNodesCostComponent(location *string, nonAzureConfigNodeCount *int64) []*schema.CostComponent {
	var nonAzureConfigNodeCountDec *decimal.Decimal

	if nonAzureConfigNodeCount != nil {
		nonAzureConfigNodeCountDec = decimalPtr(decimal.NewFromInt(*nonAzureConfigNodeCount))
	}

	costComponents := make([]*schema.CostComponent, 0)
	costComponents = append(costComponents, nonautomationDSCNodesCostComponent(*location, "5", "Non-Azure Node", "Non-Azure", nonAzureConfigNodeCountDec))

	return costComponents
}

func nonautomationDSCNodesCostComponent(location, startUsage, meterName, skuName string, monthlyQuantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{

		Name:            "Non-azure config nodes",
		Unit:            "nodes",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: monthlyQuantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(location),
			Service:       strPtr("Automation"),
			ProductFamily: strPtr("Management and Governance"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", ValueRegex: strPtr(fmt.Sprintf("/^%s$/i", meterName))},
				{Key: "skuName", ValueRegex: strPtr(fmt.Sprintf("/^%s$/i", skuName))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr(startUsage),
		},
		UsageBased: true,
	}
}
