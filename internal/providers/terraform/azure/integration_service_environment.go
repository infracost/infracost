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
	region := d.Region

	productName := "Logic Apps Integration Service Environment"
	skuName := d.Get("sku_name").String()
	sku := strings.ToLower(skuName[:strings.IndexByte(skuName, '_')])
	scaleNumber, _ := strconv.Atoi(skuName[strings.IndexByte(skuName, '_')+1:])

	costComponents := make([]*schema.CostComponent, 0)

	if sku == "developer" {
		productName += " - Developer"
	}

	costComponents = append(costComponents, IntegrationBaseServiceEnvironmentCostComponent("Base units", region, productName))

	if sku == "premium" && scaleNumber > 0 {
		costComponents = append(costComponents, IntegrationScaleServiceEnvironmentCostComponent("Scale units", region, productName, scaleNumber))

	}
	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func IntegrationBaseServiceEnvironmentCostComponent(name, region, productName string) *schema.CostComponent {
	return &schema.CostComponent{

		Name:           name,
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Logic Apps"),
			ProductFamily: strPtr("Integration"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr(productName)},
				{Key: "skuName", Value: strPtr("Base")},
				{Key: "meterName", Value: strPtr("Base Unit")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
func IntegrationScaleServiceEnvironmentCostComponent(name, region, productName string, scaleNumber int) *schema.CostComponent {
	return &schema.CostComponent{

		Name:           name,
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(int64(scaleNumber))),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
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
