package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type DNSRecordSet struct {
	Address        string
	MonthlyQueries *int64 `infracost_usage:"monthly_queries"`
}

func (r *DNSRecordSet) CoreType() string {
	return "DNSRecordSet"
}

func (r *DNSRecordSet) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_queries", ValueType: schema.Int64, DefaultValue: 0},
	}
}

func (r *DNSRecordSet) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *DNSRecordSet) BuildResource() *schema.Resource {
	var monthlyQueries *decimal.Decimal

	if r.MonthlyQueries != nil {
		monthlyQueries = decimalPtr(decimal.NewFromInt(*r.MonthlyQueries))
	}

	return &schema.Resource{
		Name: r.Address,
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
				UsageBased: true,
			},
		},
		UsageSchema: r.UsageSchema(),
	}
}
