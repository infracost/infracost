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

var AppServiceEnvironmentUsageSchema = []*schema.UsageItem{{Key: "operating_system", ValueType: schema.String, DefaultValue: "linux"}}

func (r *AppServiceEnvironment) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *AppServiceEnvironment) BuildResource() *schema.Resource {
	region := r.Region

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
	if doesStrSliceContains(stampFeeTiers, tier) == bool(true) {
		costComponents = append(costComponents, AppIsolatedServicePlanCostComponentStampFee(region, productName))
	}
	costComponents = append(costComponents, AppIsolatedServicePlanCostComponent(fmt.Sprintf("Instance usage (%s)", tier), region, productName, tier))

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents, UsageSchema: AppServiceEnvironmentUsageSchema,
	}
}
func AppIsolatedServicePlanCostComponentStampFee(region, productName string) *schema.CostComponent {
	return &schema.CostComponent{

		Name:           "Stamp fee",
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
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
func AppIsolatedServicePlanCostComponent(name, region, productName, tier string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           name,
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
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
