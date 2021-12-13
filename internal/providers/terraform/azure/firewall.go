package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetAzureRMFirewallRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_firewall",
		RFunc: NewAzureRMFirewall,
	}
}

func NewAzureRMFirewall(ctx *config.RunContext, d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := lookupRegion(d, []string{})

	var costComponents []*schema.CostComponent

	skuTier := "Standard"
	if d.Get("sku_tier").Type != gjson.Null {
		skuTier = d.Get("sku_tier").String()
	}

	// Compare d.Get() with empty array because by default an empty array of virtual hub block: "virtual_hub":[] exists,
	// and it means that d.Get("virtual_hub").Type will never return gjson.Null

	if v := d.Get("virtual_hub").String(); v != "[]" {
		if strings.ToLower(skuTier) == "standard" {
			skuTier = "Secured Virtual Hub"
		} else {
			skuTier = fmt.Sprintf("%s Secured Virtual Hub", skuTier)
		}
	}

	costComponents = append(costComponents, &schema.CostComponent{
		Name:           fmt.Sprintf("Deployment (%s)", skuTier),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
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
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: dataProcessed,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
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
