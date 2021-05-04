package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	log "github.com/sirupsen/logrus"

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

	switch sku {
	case "PC2":
		skuRefactor = "PC2"
		productName = "Premium Windows Container Plan"
	case "PC3":
		skuRefactor = "PC3"
		productName = "Premium Windows Container Plan"
	case "PC4":
		skuRefactor = "PC4"
		productName = "Premium Windows Container Plan"

	case "Y1":
		skuRefactor = "Shared"
		productName = "Shared Plan"
	case "S1":
		skuRefactor = "S1"
	case "S2":
		skuRefactor = "S2"
	case "S3":
		skuRefactor = "S3"
	case "B1":
		skuRefactor = "B1"
		productName = "Basic Plan"
	case "B2":
		skuRefactor = "B2"
		productName = "Basic Plan"
	case "B3":
		skuRefactor = "B3"
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
	if sku[:2] == "PC" && os == "linux" {
		log.Warnf("Skipping resource %s.This tariff plan is not supported for the Linux operating system", d.Address)
		return nil
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
