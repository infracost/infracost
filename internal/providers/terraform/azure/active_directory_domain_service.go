package azure

import (
	"fmt"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetAzureRMActiveDirectoryDomainServiceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_active_directory_domain_service",
		RFunc: NewAzureRMActiveDirectoryDomainService,
	}
}

func NewAzureRMActiveDirectoryDomainService(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := lookupRegion(d, []string{})
	costComponents := activeDirectoryDomainServiceCostComponents("Active directory domain service", region, d)

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func activeDirectoryDomainServiceCostComponents(name, region string, d *schema.ResourceData) []*schema.CostComponent {
	productType := "Standard"

	if d.Get("sku").Type != gjson.Null {
		productType = d.Get("sku").String()
	}

	costComponents := []*schema.CostComponent{
		{
			Name:           fmt.Sprintf("%s (%s)", name, productType),
			Unit:           "hours",
			UnitMultiplier: decimal.NewFromInt(1),
			HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("azure"),
				Region:        strPtr(region),
				Service:       strPtr("Azure Active Directory Domain Services"),
				ProductFamily: strPtr("Security"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "skuName", Value: strPtr(productType)},
					{Key: "meterName", Value: strPtr(fmt.Sprintf("%s User Forest", productType))},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("Consumption"),
			},
		},
	}
	return costComponents
}
