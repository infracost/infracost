package azure

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetAzureRMPublicIPPrefixRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_public_ip_prefix",
		RFunc: NewAzureRMPublicIPPrefix,
	}
}

func NewAzureRMPublicIPPrefix(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	location := d.Get("location").String()

	costComponents := make([]*schema.CostComponent, 0)

	costComponents = append(costComponents, PublicIPPrefixCostComponent("IP prefix", location))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
func PublicIPPrefixCostComponent(name, location string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           name,
		Unit:           "hours",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(location),
			Service:       strPtr("Virtual Network"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Public IP Prefix")},
				{Key: "meterName", Value: strPtr("Static IP Addresses")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
