package azure

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"

	"github.com/shopspring/decimal"
)

func GetAzureRMAppNATGatewayRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_nat_gateway",
		RFunc: NewAzureRMNATGateway,
	}
}

func NewAzureRMNATGateway(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var monthlyDataProcessedGb *decimal.Decimal

	if u != nil && u.Get("monthly_data_processed_gb").Type != gjson.Null {
		monthlyDataProcessedGb = decimalPtr(decimal.NewFromFloat(u.Get("monthly_data_processed_gb").Float()))
	}

	costComponents := make([]*schema.CostComponent, 0)

	costComponents = append(costComponents, NATGatewayCostComponent("NAT gateway"))
	costComponents = append(costComponents, DataProcessedCostComponent("Data processed", monthlyDataProcessedGb))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func NATGatewayCostComponent(name string) *schema.CostComponent {
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
				{Key: "productName", Value: strPtr("NAT Gateway")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
func DataProcessedCostComponent(name string, monthlyDataProcessedGb *decimal.Decimal) *schema.CostComponent {
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
				{Key: "productName", Value: strPtr("NAT Gateway")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
