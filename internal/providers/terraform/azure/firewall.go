package azure

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetAzureRMFirewallRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_firewall",
		RFunc: NewAzureFirewall,
	}
}

func NewAzureFirewall(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var costComponents []*schema.CostComponent
	location := d.Get("location").String()

	skuTier := "Standard"
	if d.Get("sku_tier").Type != gjson.Null {
		skuTier = d.Get("sku_tier").String()
	}

	if v := d.Get("virtual_hub").String(); v != "[]" {
		if skuTier == "Standard" {
			skuTier = "Secured Virtual Hub"
		} else {
			skuTier = fmt.Sprintf("%s Secured Virtual Hub", skuTier)
		}
	}

	costComponents = append(costComponents, &schema.CostComponent{
		Name:           fmt.Sprintf("Deployment (%s)", skuTier),
		Unit:           "hours",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(location),
			Service:       strPtr("Azure Firewall"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", Value: strPtr(fmt.Sprintf("%s Deployment", skuTier))},
				{Key: "skuName", Value: strPtr(skuTier)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	})

	var dataProcessed *decimal.Decimal
	if u != nil && u.Get("monthly_data_processed_gb").Exists() {
		dataProcessed = decimalPtr(decimal.NewFromInt(u.Get("monthly_data_processed_gb").Int()))
	}

	costComponents = append(costComponents, &schema.CostComponent{
		Name:            "Data processed",
		Unit:            "GB",
		UnitMultiplier:  1,
		MonthlyQuantity: dataProcessed,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(location),
			Service:       strPtr("Azure Firewall"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", Value: strPtr(fmt.Sprintf("%s Data Processed", skuTier))},
				{Key: "skuName", Value: strPtr(skuTier)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	})

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
