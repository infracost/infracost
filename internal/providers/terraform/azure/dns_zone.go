package azure

import (
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetAzureRMDNSZoneRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_dns_zone",
		RFunc: NewAzureRMDNSZone,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
		Notes: []string{"Most expensive price tier is used."},
	}
}

func NewAzureRMDNSZone(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	group := d.References("resource_group_name")
	location := group[0].Get("location").String()

	if strings.HasPrefix(strings.ToLower(location), "usgov") {
		location = "US Gov Zone 1"
	}
	if strings.HasPrefix(strings.ToLower(location), "germany") {
		location = "DE Zone 1"
	}
	if strings.HasPrefix(strings.ToLower(location), "china") {
		location = "Zone 1 (China)"
	}
	if location != "US Gov Zone 1" && location != "DE Zone 1" && location != "Zone 1 (China)" {
		location = "Zone 1"
	}

	costComponents := make([]*schema.CostComponent, 0)

	costComponents = append(costComponents, hostedPublicZoneCostComponent(location))
	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
func hostedPublicZoneCostComponent(region string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Hosted zone",
		Unit:            "months",
		UnitMultiplier:  1,
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Azure DNS"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", Value: strPtr("Public Zones")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr("0"),
		},
	}
}
