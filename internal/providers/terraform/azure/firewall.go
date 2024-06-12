package azure

import (
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMFirewallRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_firewall",
		RFunc: NewAzureRMFirewall,
	}
}

func NewAzureRMFirewall(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Region

	var costComponents []*schema.CostComponent

	skuTier := "Standard"
	if d.Get("sku_tier").Type != gjson.Null {
		skuTier = cases.Title(language.English).String(d.Get("sku_tier").String())
	}

	if len(d.Get("virtual_hub").Array()) > 0 {
		if skuTier == "Standard" {
			skuTier = "Standard Secure Virtual Hub"
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
