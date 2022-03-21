package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetAzureRMAppServicePlanRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_app_service_plan",
		RFunc: NewAzureRMAppServicePlan,
	}
}

func NewAzureRMAppServicePlan(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := lookupRegion(d, []string{})

	sku := d.Get("sku.0.size").String()
	skuRefactor := ""
	os := "windows"
	capacity := d.GetInt64OrDefault("sku.0.capacity", 1)
	productName := "Standard Plan"

	// These are used by azurerm_function_app, their costs are calculated there as they don't have prices in the azurerm_app_service_plan resource
	if len(sku) < 2 || strings.ToLower(sku[:2]) == "ep" {
		return &schema.Resource{
			Name:      d.Address,
			IsSkipped: true,
			NoPrice:   true,
		}
	}

	switch strings.ToLower(sku[2:]) {
	case "v1":
		skuRefactor = sku[:2]
		productName = "Premium Plan"
	case "v2":
		skuRefactor = sku[:2] + " " + sku[2:]
		productName = "Premium v2 Plan"
	case "v3":
		skuRefactor = sku[:2] + " " + sku[2:]
		productName = "Premium v3 Plan"
	}

	switch strings.ToLower(sku[:2]) {
	case "pc":
		skuRefactor = "PC" + sku[2:]
		productName = "Premium Windows Container Plan"
	case "y1":
		skuRefactor = "Shared"
		productName = "Shared Plan"
	}

	switch strings.ToLower(sku[:1]) {
	case "s":
		skuRefactor = "S" + sku[1:]
	case "b":
		skuRefactor = "B" + sku[1:]
		productName = "Basic Plan"
	}

	if d.Get("kind").Exists() {
		os = strings.ToLower(d.Get("kind").String())
	}
	if os == "app" {
		os = "windows"
	}
	if os != "windows" && productName != "Premium Plan" {
		productName += " - Linux"
	}

	costComponents := make([]*schema.CostComponent, 0)

	costComponents = append(costComponents, AppServicePlanCostComponent(fmt.Sprintf("Instance usage (%s)", sku), region, productName, skuRefactor, capacity))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func AppServicePlanCostComponent(name, region, productName, skuRefactor string, capacity int64) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           name,
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(capacity)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Azure App Service"),
			ProductFamily: strPtr("Compute"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Azure App Service " + productName)},
				{Key: "skuName", ValueRegex: strPtr(fmt.Sprintf("/%s/i", skuRefactor))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
