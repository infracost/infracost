package azure

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

func GetAzureRMAppServicePlanRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_app_service_plan",
		RFunc: NewAzureRMAppServicePlan,
		Notes: []string{
			"Costs associated with running an app service plan in Azure",
		},
	}
}

func NewAzureRMAppServicePlan(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	return &schema.Resource{
		Name:           d.Address,
		CostComponents: AppServicePlanCostComponent(d),
	}
}

func AppServicePlanCostComponent(d *schema.ResourceData) []*schema.CostComponent {
	sku := d.Get("sku.0.size").String()
	skuRefactor := ""
	purchaseOption := "Consumption"
	capacity := d.Get("sku.0.capacity").Int()
	location := d.Get("location").String()
	productName := "Standard Plan"

	if sku == "P1v2" {
		skuRefactor = "P1 v2"
	}
	if sku == "P2v2" {
		skuRefactor = "P2 v2"
	}
	if sku == "P3v2" {
		skuRefactor = "P3 v2"
	}
	if sku == "P1v3" {
		skuRefactor = "P1 v3"
	}
	if sku == "P2v3" {
		skuRefactor = "P2 v3"
	}
	if sku == "P3v3" {
		skuRefactor = "P3 v3"
	}
	if sku == "Y1" {
		skuRefactor = "Shared"
	}
	if sku == "S1" {
		skuRefactor = "S1"
	}
	if sku == "S2" {
		skuRefactor = "S2"
	}
	if sku == "S3" {
		skuRefactor = "S3"
	}
	if sku == "B1" {
		skuRefactor = "B1"
	}
	if sku == "B2" {
		skuRefactor = "B2"
	}
	if sku == "B3" {
		skuRefactor = "B3"
	}

	switch skuRefactor {
	case "P1 v2":
		productName = "Premium v2 Plan"
	case "P2 v2":
		productName = "Premium v2 Plan"
	case "P3 v2":
		productName = "Premium v2 Plan"
	case "P1 v3":
		productName = "Premium v3 Plan"
	case "P2 v3":
		productName = "Premium v3 Plan"
	case "P3 v3":
		productName = "Premium v3 Plan"
	case "Shared":
		productName = "Shared Plan"
	case "B1":
		productName = "Basic Plan"
	case "B2":
		productName = "Basic Plan"
	case "B3":
		productName = "Basic Plan"
	}

	os := "Windows"
	if d.Get("kind").Exists() {
		os = d.Get("kind").String()
	}
	if os != "Windows" {
		productName += " - Linux"
	}

	costComponents := make([]*schema.CostComponent, 0)

	costComponents = append(costComponents, &schema.CostComponent{
		Name:           fmt.Sprintf("Instance usage (%s)", sku),
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
			PurchaseOption: strPtr(purchaseOption),
		},
	})

	return costComponents
}
