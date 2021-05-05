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
	sku := d.Get("sku.0.size").String()
	skuRefactor := ""
	os := "Windows"
	capacity := d.Get("sku.0.capacity").Int()
	location := d.Get("location").String()
	productName := "Standard Plan"

	switch sku[2:] {
	case "v2":
		skuRefactor = sku[:2] + " " + sku[2:]
		productName = "Premium v2 Plan"
	case "v3":
		skuRefactor = sku[:2] + " " + sku[2:]
		productName = "Premium v3 Plan"
	}

	switch sku[:2] {
	case "PC":
		skuRefactor = "PC" + sku[2:]
		productName = "Premium Windows Container Plan"
	case "Y1":
		skuRefactor = "Shared"
		productName = "Shared Plan"
	}

	switch sku[:1] {
	case "S":
		skuRefactor = "S" + sku[1:]
	case "B":
		skuRefactor = "B" + sku[1:]
		productName = "Basic Plan"
	}

	if d.Get("kind").Exists() {
		os = strings.ToLower(d.Get("kind").String())
	}
	if os == "app" {
		os = "windows"
	}
	if os != "windows" {
		productName += " - Linux"
	}

	costComponents := make([]*schema.CostComponent, 0)

	costComponents = append(costComponents, AppServicePlanCostComponent(fmt.Sprintf("Instance usage (%s)", sku), location, productName, skuRefactor, capacity))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func AppServicePlanCostComponent(name, location, productName, skuRefactor string, capacity int64) *schema.CostComponent {

	return &schema.CostComponent{

		Name:           name,
		Unit:           "hours",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(capacity)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(location),
			Service:       strPtr("Azure App Service"),
			ProductFamily: strPtr("Compute"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Azure App Service " + productName)},
				{Key: "skuName", Value: strPtr(skuRefactor)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
