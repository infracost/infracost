package azure

import (
	"fmt"
	"strings"

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

	group := d.References("resource_group_name")[0]

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: NodesCostComponent(d, u, group),
	}
}

func NodesCostComponent(d *schema.ResourceData, u *schema.UsageData, group *schema.ResourceData) []*schema.CostComponent {
	var nonAzureConfigNodeCount *decimal.Decimal

	location := strings.ToLower(group.Get("location").String())

	if u != nil && u.Get("non_azure_config_node_count").Type != gjson.Null {
		nonAzureConfigNodeCount = decimalPtr(decimal.NewFromInt(u.Get("non_azure_config_node_count").Int()))
	}

	costComponents := make([]*schema.CostComponent, 0)

	costComponents = append(costComponents, NonNodesCostComponent(location, "5", "Non-Azure Node", nonAzureConfigNodeCount))

	return costComponents
}

func NonNodesCostComponent(location, startUsage, meterName string, monthlyQuantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{

		Name:            "Non-azure config nodes",
		Unit:            "nodes",
		UnitMultiplier:  1,
		MonthlyQuantity: monthlyQuantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(location),
			Service:       strPtr("Automation"),
			ProductFamily: strPtr("Management and Governance"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", ValueRegex: strPtr(fmt.Sprintf("/^%s$/i", meterName))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr(startUsage),
		},
	}
}
