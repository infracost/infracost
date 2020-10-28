package aws

import (
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

func GetRoute53ZoneRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_route53_zone",
		RFunc: NewRoute53Zone,
	}
}

func NewRoute53Zone(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Per hosted zone",
				Unit:            "months",
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
	}
}
