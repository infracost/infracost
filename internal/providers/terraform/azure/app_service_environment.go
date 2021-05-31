package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	log "github.com/sirupsen/logrus"
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
	tier := "I1"
	if d.Get("pricing_tier").Type != gjson.Null {
		tier = d.Get("pricing_tier").String()
	}
	location := ""
	group := d.References("resource_group_name")
	if len(group) > 0 {
		location = group[0].Get("location").String()
	}
	if location == "" {
		log.Warnf("Skipping resource %s. Could not find its 'location' property.", d.Address)
		return nil
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
		costComponents = append(costComponents, AppIsolatedServicePlanCostComponentStampFee(location, productName))
	}
	costComponents = append(costComponents, AppIsolatedServicePlanCostComponent(fmt.Sprintf("Instance usage (%s)", tier), location, productName, tier))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
func AppIsolatedServicePlanCostComponentStampFee(location, productName string) *schema.CostComponent {
	return &schema.CostComponent{

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
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
func AppIsolatedServicePlanCostComponent(name, location, productName, tier string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           name,
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
