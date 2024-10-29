package azure

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type AppServiceEnvironment struct {
	Address         string
	Region          string
	PricingTier     string
	OperatingSystem *string `infracost_usage:"operating_system"`
}

func (r *AppServiceEnvironment) CoreType() string {
	return "AppServiceEnvironment"
}

func (r *AppServiceEnvironment) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{{Key: "operating_system", ValueType: schema.String, DefaultValue: "linux"}}
}

func (r *AppServiceEnvironment) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *AppServiceEnvironment) BuildResource() *schema.Resource {
	tier := "I1"
	if r.PricingTier != "" {
		tier = r.PricingTier
	}

	stampFeeTiers := []string{"I1", "I2", "I3"}
	productName := "Isolated Plan"
	costComponents := make([]*schema.CostComponent, 0)
	os := "linux"
	if r.OperatingSystem != nil {
		os = strings.ToLower(*r.OperatingSystem)
	}
	if os == "linux" {
		productName += " - Linux"
	}
	if contains(stampFeeTiers, tier) == bool(true) {
		costComponents = append(costComponents, r.appIsolatedServicePlanCostComponentStampFee(productName))
	}
	costComponents = append(costComponents, r.appIsolatedServicePlanCostComponent(fmt.Sprintf("Instance usage (%s)", tier), productName, tier))

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}
func (r *AppServiceEnvironment) appIsolatedServicePlanCostComponentStampFee(productName string) *schema.CostComponent {
	return &schema.CostComponent{

		Name:           "Stamp fee",
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Azure App Service"),
			ProductFamily: strPtr("Compute"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Azure App Service " + productName)},
				{Key: "skuName", Value: strPtr("Stamp")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
func (r *AppServiceEnvironment) appIsolatedServicePlanCostComponent(name, productName, tier string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           name,
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Azure App Service"),
			ProductFamily: strPtr("Compute"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Azure App Service " + productName)},
				{Key: "skuName", Value: strPtr(tier)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
