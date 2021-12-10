package aws

import (
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

func GetRoute53ZoneRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_route53_zone",
		RFunc: NewRoute53Zone,
	}
}

func NewRoute53Zone(ctx *config.ProjectContext, d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	return &schema.Resource{
		Name: d.Address,
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
	}
}
