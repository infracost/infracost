package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

func GetAzureRMAppIsolatedServicePlanRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_app_service_environment",
		RFunc: NewAzureRMAppIsolatedServicePlan,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
		Notes: []string{
			"Costs associated with running an app service plan in Azure",
		},
	}
}

func NewAzureRMAppIsolatedServicePlan(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	return &schema.Resource{
		Name:           d.Address,
		CostComponents: AppIsolatedServicePlanCostComponent(d, u),
	}
}
func AppIsolatedServicePlanCostComponent(d *schema.ResourceData, u *schema.UsageData) []*schema.CostComponent {

	tier := d.Get("pricing_tier").String()
	group := d.References("resource_group_name")
	location := group[0].Get("location").String()
	stampFeeTiers := []string{"I1", "I2", "I3"}
	productName := "Isolated Plan"
	purchaseOption := "Consumption"
	costComponents := make([]*schema.CostComponent, 0)
	os := "Linux"
	if u != nil && u.Get("operating_system").Exists() {
		os = strings.ToLower(u.Get("operating_system").String())
	}
	if os == "Linux" {
		productName += " - Linux"
	}

	if Contains(stampFeeTiers, tier) == bool(true) {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:           "Stamp fee",
			Unit:           "hours",
			UnitMultiplier: 1,
			HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("azure"),
				Region:        strPtr(location),
				Service:       strPtr("Azure App Service"),
				ProductFamily: strPtr("Compute"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Azure App Service " + productName)},
					{Key: "skuName", Value: strPtr("Stamp")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr(purchaseOption),
			},
		})
	}

	costComponents = append(costComponents, &schema.CostComponent{
		Name:           fmt.Sprintf("Instance usage (%s)", tier),
		Unit:           "hours",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(location),
			Service:       strPtr("Azure App Service"),
			ProductFamily: strPtr("Compute"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Azure App Service " + productName)},
				{Key: "skuName", Value: strPtr(tier)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr(purchaseOption),
		},
	})

	return costComponents

}

func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
