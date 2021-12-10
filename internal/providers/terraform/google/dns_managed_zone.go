package google

import (
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetDNSManagedZoneRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_dns_managed_zone",
		RFunc: NewDNSManagedZone,
	}
}

func NewDNSManagedZone(ctx *config.ProjectContext, d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Managed zone",
				Unit:            "months",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("gcp"),
					Region:        strPtr("global"),
					Service:       strPtr("Cloud DNS"),
					ProductFamily: strPtr("Network"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "description", Value: strPtr("ManagedZone")},
					},
				},
				PriceFilter: &schema.PriceFilter{
					StartUsageAmount: strPtr("0"),
				},
			},
		},
	}
}
