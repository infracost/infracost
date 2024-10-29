package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type DNSManagedZone struct {
	Address string
}

func (r *DNSManagedZone) CoreType() string {
	return "DNSManagedZone"
}

func (r *DNSManagedZone) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *DNSManagedZone) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *DNSManagedZone) BuildResource() *schema.Resource {
	return &schema.Resource{
		Name: r.Address,
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
		UsageSchema: r.UsageSchema(),
	}
}
