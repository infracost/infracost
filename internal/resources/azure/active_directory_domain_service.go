package azure

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"

	"github.com/shopspring/decimal"
)

type ActiveDirectoryDomainService struct {
	Address string
	Region  string
	SKU     string
}

func (r *ActiveDirectoryDomainService) CoreType() string {
	return "ActiveDirectoryDomainService"
}

func (r *ActiveDirectoryDomainService) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *ActiveDirectoryDomainService) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ActiveDirectoryDomainService) BuildResource() *schema.Resource {
	costComponents := activeDirectoryDomainServiceCostComponents("Active directory domain service", r.Region, r.SKU)

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func activeDirectoryDomainServiceCostComponents(name, region, sku string) []*schema.CostComponent {
	productType := "Standard"

	if sku != "" {
		productType = sku
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
				Service:       strPtr("Microsoft Entra Domain Services"),
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
