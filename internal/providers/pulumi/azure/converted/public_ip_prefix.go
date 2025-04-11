package azure

import (
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMPublicIPPrefixRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_public_ip_prefix",
		RFunc: NewAzureRMPublicIPPrefix,
	}
}

func NewAzureRMPublicIPPrefix(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Region

	costComponents := make([]*schema.CostComponent, 0)

	costComponents = append(costComponents, PublicIPPrefixCostComponent("IP prefix", region))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
func PublicIPPrefixCostComponent(name, region string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           name,
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Virtual Network"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Public IP Prefix")},
				{Key: "meterName", ValueRegex: strPtr("/Static IP Addresses/i")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
