package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"

	"github.com/shopspring/decimal"
)

func GetAzureRMAppIsolatedServicePlanRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_app_service_environment",
		RFunc: NewAzureRMAppIsolatedServicePlan,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func NewAzureRMAppIsolatedServicePlan(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := lookupRegion(d, []string{"resource_group_name"})

	tier := "I1"
	if d.Get("pricing_tier").Type != gjson.Null {
		tier = d.Get("pricing_tier").String()
	}

	stampFeeTiers := []string{"I1", "I2", "I3"}
	productName := "Isolated Plan"
	costComponents := make([]*schema.CostComponent, 0)
	os := "linux"
	if u != nil && u.Get("operating_system").Type != gjson.Null {
		os = strings.ToLower(u.Get("operating_system").String())
	}
	if os == "linux" {
		productName += " - Linux"
	}
	if Contains(stampFeeTiers, tier) == bool(true) {
		costComponents = append(costComponents, AppIsolatedServicePlanCostComponentStampFee(region, productName))
	}
	costComponents = append(costComponents, AppIsolatedServicePlanCostComponent(fmt.Sprintf("Instance usage (%s)", tier), region, productName, tier))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
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

func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
