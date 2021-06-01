package azure

import (
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetAzureRMAppNATGatewayRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_nat_gateway",
		RFunc: NewAzureRMNATGateway,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func NewAzureRMNATGateway(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var monthlyDataProcessedGb *decimal.Decimal
	location := "Global"
	group := d.References("resource_group_name")

	if len(group) > 0 {
		if strings.HasPrefix(strings.ToLower(group[0].Get("location").String()), "usgov") {
			location = "US Gov"
		}
		if strings.Contains(strings.ToLower(group[0].Get("location").String()), "china") {
			location = "Сhina"
		}
	}

	if u != nil && u.Get("monthly_data_processed_gb").Type != gjson.Null {
		monthlyDataProcessedGb = decimalPtr(decimal.NewFromFloat(u.Get("monthly_data_processed_gb").Float()))
	}

	costComponents := make([]*schema.CostComponent, 0)

	costComponents = append(costComponents, NATGatewayCostComponent("NAT gateway", location))
	costComponents = append(costComponents, DataProcessedCostComponent("Data processed", location, monthlyDataProcessedGb))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func NATGatewayCostComponent(name, location string) *schema.CostComponent {
	return &schema.CostComponent{

		Name:           name,
		Unit:           "hours",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr("Global"),
			Service:       strPtr("NAT Gateway"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", Value: strPtr("Gateway")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
func DataProcessedCostComponent(name, location string, monthlyDataProcessedGb *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{

		Name:            name,
		Unit:            "GB",
		UnitMultiplier:  1,
		MonthlyQuantity: monthlyDataProcessedGb,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr("Global"),
			Service:       strPtr("NAT Gateway"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", Value: strPtr("Data Processed")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
