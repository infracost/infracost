package azure

import (
	"strconv"
	"strings"

	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

func GetAzureRMAppIntegrationServiceEnvironmentRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_integration_service_environment",
		RFunc: NewAzureRMIntegrationServiceEnvironment,
	}
}

func NewAzureRMIntegrationServiceEnvironment(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	productName := "Logic Apps Integration Service Environment"
	location := d.Get("location").String()
	sku_name := d.Get("sku_name").String()
	sku := strings.ToLower(sku_name[:strings.IndexByte(sku_name, '_')])
	scaleNumber, _ := strconv.Atoi(sku_name[strings.IndexByte(sku_name, '_')+1:])

	costComponents := make([]*schema.CostComponent, 0)

	if sku == "developer" {
		productName = productName + " - Developer"
	}

	costComponents = append(costComponents, IntegrationBaseServiceEnvironmentCostComponent("Base units", location, productName))

	if sku == "premium" && scaleNumber > 0 {
		costComponents = append(costComponents, IntegrationScaleServiceEnvironmentCostComponent("Scale units", location, productName, scaleNumber))

	}
	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func IntegrationBaseServiceEnvironmentCostComponent(name, location, productName string) *schema.CostComponent {
	return &schema.CostComponent{

		Name:           name,
		Unit:           "hours",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(location),
			Service:       strPtr("Logic Apps"),
			ProductFamily: strPtr("Integration"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr(productName)},
				{Key: "skuName", Value: strPtr("Base")},
				{Key: "meterName", Value: strPtr("Base Units")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
func IntegrationScaleServiceEnvironmentCostComponent(name, location, productName string, scaleNumber int) *schema.CostComponent {
	return &schema.CostComponent{

		Name:           name,
		Unit:           "hours",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(int64(scaleNumber))),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(location),
			Service:       strPtr("Logic Apps"),
			ProductFamily: strPtr("Integration"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr(productName)},
				{Key: "skuName", Value: strPtr("Scale")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
