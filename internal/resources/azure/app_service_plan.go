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
	sku := ""
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

	switch strings.ToLower(r.SKUSize[2:]) {
	case "v1":
		sku = r.SKUSize[:2]
		productName = "Premium Plan"
	case "v2":
		sku = r.SKUSize[:2] + " " + r.SKUSize[2:]
		productName = "Premium v2 Plan"
	case "v3":
		sku = r.SKUSize[:2] + " " + r.SKUSize[2:]
		productName = "Premium v3 Plan"
	}

	switch strings.ToLower(r.SKUSize[:2]) {
	case "pc":
		sku = "PC" + r.SKUSize[2:]
		productName = "Premium Windows Container Plan"
	case "y1":
		sku = "Shared"
		productName = "Shared Plan"
	}

	switch strings.ToLower(r.SKUSize[:1]) {
	case "s":
		sku = "S" + r.SKUSize[1:]
	case "b":
		sku = "B" + r.SKUSize[1:]
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

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: []*schema.CostComponent{r.appServicePlanCostComponent(fmt.Sprintf("Instance usage (%s)", r.SKUSize), productName, sku, capacity)},
		UsageSchema:    AppServicePlanUsageSchema,
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
