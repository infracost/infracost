package azure

import (
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
	region := lookupRegion(d, []string{"resource_group_name"})
	region = convertRegion(region)

	var monthlyDataProcessedGb *decimal.Decimal
	if u != nil && u.Get("monthly_data_processed_gb").Type != gjson.Null {
		monthlyDataProcessedGb = decimalPtr(decimal.NewFromFloat(u.Get("monthly_data_processed_gb").Float()))
	}

	costComponents := make([]*schema.CostComponent, 0)

	costComponents = append(costComponents, NATGatewayCostComponent("NAT gateway", region))
	costComponents = append(costComponents, DataProcessedCostComponent("Data processed", region, monthlyDataProcessedGb))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func NATGatewayCostComponent(name, region string) *schema.CostComponent {
	return &schema.CostComponent{

		Name:           name,
		Unit:           "hours",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
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
func DataProcessedCostComponent(name, region string, monthlyDataProcessedGb *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{

		Name:            name,
		Unit:            "GB",
		UnitMultiplier:  1,
		MonthlyQuantity: monthlyDataProcessedGb,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
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
