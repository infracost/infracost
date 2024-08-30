package azure

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMPublicIPRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_public_ip",
		RFunc: NewAzureRMPublicIP,
	}
}

func NewAzureRMPublicIP(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Region

	var meterName string
	sku := "Standard" // default sku is Standard
	skuTier := "Regional"
	allocationMethod := d.Get("allocation_method").String()

	if d.Get("sku").Type != gjson.Null {
		sku = d.Get("sku").String()
	}

	switch sku {
	case "Basic":
		meterName = "Basic IPv4 " + allocationMethod + " Public IP"
	case "Standard":
		skuTierVal := d.Get("sku_tier").String()
		if skuTierVal != "" {
			skuTier = skuTierVal
		}
		if skuTier == "Global" {
			sku = "Standard" // When sku_tier is Global, sku is Standard
			meterName = "Global IPv4 " + allocationMethod + " Public IP"
		} else {
			meterName = "Standard IPv4 " + allocationMethod + " Public IP"
		}
	}

	name := fmt.Sprintf("IP address (%s, %s)", strings.ToLower(allocationMethod), strings.ToLower(skuTier))

	costComponents := make([]*schema.CostComponent, 0)

	costComponents = append(costComponents, PublicIPCostComponent(name, region, sku, meterName))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
func PublicIPCostComponent(name, region, sku, meterName string) *schema.CostComponent {
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
				{Key: "productName", Value: strPtr("IP Addresses")},
				{Key: "skuName", Value: strPtr(sku)},
				{Key: "meterName", Value: strPtr(meterName)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
