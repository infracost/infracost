package google

import (
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetDNSRecordSetRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_dns_record_set",
		RFunc: NewDNSRecordSet,
	}
}

func NewDNSRecordSet(ctx *config.RunContext, d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var monthlyQueries *decimal.Decimal

	if u != nil && u.Get("monthly_queries").Exists() {
		monthlyQueries = decimalPtr(decimal.NewFromInt(u.Get("monthly_queries").Int()))
	}

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Queries",
				Unit:            "1M queries",
				UnitMultiplier:  decimal.NewFromInt(1000000),
				MonthlyQuantity: monthlyQueries,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("gcp"),
					Region:        strPtr("global"),
					Service:       strPtr("Cloud DNS"),
					ProductFamily: strPtr("Network"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "description", Value: strPtr("DNS Query (port 53)")},
					},
				},
				PriceFilter: &schema.PriceFilter{
					StartUsageAmount: strPtr("0"),
				},
			},
		},
	}
}
