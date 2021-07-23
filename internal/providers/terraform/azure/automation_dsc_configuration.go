package azure

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetAzureRMAutomationDscConfigurationRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_automation_dsc_configuration",
		RFunc: NewAzureRMAutomationDscConfiguration,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func NewAzureRMAutomationDscConfiguration(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	return &schema.Resource{
		Name:           d.Address,
		CostComponents: nodesCostComponent(d, u),
	}
}

func nodesCostComponent(d *schema.ResourceData, u *schema.UsageData) []*schema.CostComponent {
	var nonAzureConfigNodeCount *decimal.Decimal
	location := lookupRegion(d, []string{"resource_group_name"})

	if u != nil && u.Get("non_azure_config_node_count").Type != gjson.Null {
		nonAzureConfigNodeCount = decimalPtr(decimal.NewFromInt(u.Get("non_azure_config_node_count").Int()))
	}

	costComponents := make([]*schema.CostComponent, 0)
	costComponents = append(costComponents, nonNodesCostComponent(location, "5", "Non-Azure Node", "Non-Azure", nonAzureConfigNodeCount))

	return costComponents
}

func nonNodesCostComponent(location, startUsage, meterName, skuName string, monthlyQuantity *decimal.Decimal) *schema.CostComponent {
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
	}
}
