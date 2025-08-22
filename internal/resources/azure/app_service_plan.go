package azure

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type AppServicePlan struct {
	Address     string
	SKUSize     string
	SKUCapacity int64
	Kind        string
	Region      string
	IsDevTest   bool
}

func (r *AppServicePlan) CoreType() string {
	return "AppServicePlan"
}

func (r *AppServicePlan) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *AppServicePlan) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *AppServicePlan) BuildResource() *schema.Resource {
	sku := ""
	os := "windows"
	var capacity int64 = 1
	if r.SKUCapacity > 0 {
		capacity = r.SKUCapacity
	}
	productName := "Standard Plan"

	if len(r.SKUSize) < 2 || strings.ToLower(r.SKUSize[:2]) == "ep" || strings.ToLower(r.SKUSize[:2]) == "y1" || strings.ToLower(r.SKUSize[:2]) == "ws" {
		return &schema.Resource{
			Name:        r.Address,
			IsSkipped:   true,
			NoPrice:     true,
			UsageSchema: r.UsageSchema(),
		}
	}

	var additionalAttributeFilters []*schema.AttributeFilter

	switch strings.ToLower(r.SKUSize[:1]) {
	case "s":
		sku = "S" + r.SKUSize[1:]
	case "b":
		sku = "B" + r.SKUSize[1:]
		productName = "Basic Plan"
	case "p", "i":
		sku, productName, additionalAttributeFilters = getVersionedAppServicePlanSKU(r.SKUSize, os)
	}

	switch strings.ToLower(r.SKUSize[:2]) {
	case "pc":
		sku = "PC" + r.SKUSize[2:]
		productName = "Premium Windows Container Plan"
	case "y1":
		sku = "Shared"
		productName = "Shared Plan"
	}

	if r.Kind != "" {
		os = strings.ToLower(r.Kind)
	}
	if os == "app" {
		os = "windows"
	}
	if os != "windows" && productName != "Premium Plan" && productName != "Isolated Plan" {
		productName += " - Linux"
	}

	purchaseOption := "Consumption"
	name := fmt.Sprintf("Instance usage (%s)", r.SKUSize)
	if r.IsDevTest && strings.Contains(os, "windows") && strings.ToLower(r.SKUSize[:2]) != "pc" {
		purchaseOption = "DevTestConsumption"
		name = fmt.Sprintf("Instance usage (dev/test, %s)", r.SKUSize)
	}

	return &schema.Resource{
		Name: r.Address,
		CostComponents: []*schema.CostComponent{
			servicePlanCostComponent(
				r.Region,
				name,
				productName,
				sku,
				capacity,
				purchaseOption,
				additionalAttributeFilters...,
			),
		},
		UsageSchema: r.UsageSchema(),
	}
}

func servicePlanCostComponent(region, name, productName, skuRefactor string, capacity int64, purchaseOption string, additionalAttributeFilters ...*schema.AttributeFilter) *schema.CostComponent {
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
			AttributeFilters: append([]*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Azure App Service " + productName)},
				{Key: "skuName", ValueRegex: strPtr(fmt.Sprintf("/%s$/i", skuRefactor))},
			}, additionalAttributeFilters...),
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr(purchaseOption),
		},
	}
}

func getVersionedAppServicePlanSKU(skuName, os string) (string, string, []*schema.AttributeFilter) {
	tier := "Premium"
	if strings.ToLower(skuName[:1]) == "i" {
		tier = "Isolated"
	}

	version := strings.ToLower(skuName[2:])
	if version == "v1" {
		version = ""
	}

	formattedSku := strings.TrimSpace(skuName[:2] + " ?" + version)

	productVersion := version
	if len(version) > 0 && version[0] == 'm' {
		productVersion = version[1:]
	}
	productName := strings.ReplaceAll(tier+" "+productVersion+" Plan", "  ", " ")

	if productVersion == "v3" && os == "linux" {
		return formattedSku, productName, []*schema.AttributeFilter{
			{
				Key:        "armSkuName",
				ValueRegex: strPtr(fmt.Sprintf("/%s$/i", strings.ReplaceAll(formattedSku, " ", "_"))),
			},
		}
	}

	return formattedSku, productName, nil
}
