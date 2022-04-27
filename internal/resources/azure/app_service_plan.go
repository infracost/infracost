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
}

var AppServicePlanUsageSchema = []*schema.UsageItem{}

func (r *AppServicePlan) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *AppServicePlan) BuildResource() *schema.Resource {
	sku := r.SKUSize
	skuRefactor := ""
	os := "windows"
	var capacity int64 = 1
	if r.SKUCapacity > 0 {
		capacity = r.SKUCapacity
	}
	productName := "Standard Plan"

	if len(r.SKUSize) < 2 || strings.ToLower(r.SKUSize[:2]) == "ep" {
		return &schema.Resource{
			Name:      r.Address,
			IsSkipped: true,
			NoPrice:   true, UsageSchema: AppServicePlanUsageSchema,
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

	if r.Kind != "" {
		os = strings.ToLower(r.Kind)
	}
	if os == "app" {
		os = "windows"
	}
	if os != "windows" && productName != "Premium Plan" {
		productName += " - Linux"
	}

	costComponents := make([]*schema.CostComponent, 0)

	costComponents = append(costComponents, r.appServicePlanCostComponent(fmt.Sprintf("Instance usage (%s)", sku), productName, skuRefactor, capacity))

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents, UsageSchema: AppServicePlanUsageSchema,
	}
}

func (r *AppServicePlan) appServicePlanCostComponent(name, productName, skuRefactor string, capacity int64) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           name,
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(capacity)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
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
