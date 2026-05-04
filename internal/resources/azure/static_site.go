package azure

import (
	"strings"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// StaticSite struct represents an Azure Static Web App. Static Web Apps is a service
// that automatically builds and deploys full stack web apps to Azure from a GitHub repository.
//
// Resource information: https://learn.microsoft.com/en-us/azure/static-web-apps/
// Pricing information: https://azure.microsoft.com/en-us/pricing/details/app-service/static/
type StaticSite struct {
	Address string
	Region  string
	SKU     string
}

func (r *StaticSite) CoreType() string {
	return "StaticSite"
}

func (r *StaticSite) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *StaticSite) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *StaticSite) BuildResource() *schema.Resource {
	if strings.ToLower(r.SKU) == "free" {
		return &schema.Resource{
			Name:      r.Address,
			NoPrice:   true,
			IsSkipped: true,
		}
	}

	return &schema.Resource{
		Name: r.Address,
		CostComponents: []*schema.CostComponent{
			r.staticSiteCostComponent(),
		},
		UsageSchema: r.UsageSchema(),
	}
}

func (r *StaticSite) staticSiteCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Static Web App (Standard)",
		Unit:            "months",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Azure App Service"),
			ProductFamily: strPtr("Compute"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Static Web Apps")},
				{Key: "skuName", Value: strPtr("Standard")},
				{Key: "meterName", Value: strPtr("Standard App")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
