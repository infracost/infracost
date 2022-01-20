package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type Route53Zone struct {
	Address string
}

var Route53ZoneUsageSchema = []*schema.UsageItem{}

func (r *Route53Zone) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *Route53Zone) BuildResource() *schema.Resource {
	return &schema.Resource{
		Name: r.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Hosted zone",
				Unit:            "months",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Service:       strPtr("AmazonRoute53"),
					ProductFamily: strPtr("DNS Zone"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", Value: strPtr("HostedZone")},
					},
				},
				PriceFilter: &schema.PriceFilter{
					StartUsageAmount: strPtr("0"),
				},
			},
		},
		UsageSchema: Route53ZoneUsageSchema,
	}
}
