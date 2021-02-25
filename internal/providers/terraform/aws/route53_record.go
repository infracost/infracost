package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetRoute53RecordRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_route53_record",
		RFunc:               NewRoute53Record,
		ReferenceAttributes: []string{"alias.0.name"},
	}
}

func NewRoute53Record(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	if len(d.References("alias.0.name")) > 0 && d.References("alias.0.name")[0].Type != "aws_route53_record" {
		return &schema.Resource{
			NoPrice:   true,
			IsSkipped: true,
		}
	}

	usageType := "DNS-Queries"
	usageName := "Standard queries"
	if d.Get("geolocation_routing_policy.0").Exists() {
		usageType = "Geo-Queries"
		usageName = "Geo DNS queries"
	} else if d.Get("latency_routing_policy.0").Exists() {
		usageType = "LBR-Queries"
		usageName = "Latency based routing queries"
	}

	var numberOfQueries *decimal.Decimal
	if u != nil && u.Get("monthly_queries").Exists() {
		numberOfQueries = decimalPtr(decimal.NewFromInt(u.Get("monthly_queries").Int()))
	}

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            usageName,
				Unit:            "queries",
				UnitMultiplier:  1000000,
				MonthlyQuantity: numberOfQueries,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Service:       strPtr("AmazonRoute53"),
					ProductFamily: strPtr("DNS Query"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", Value: &usageType},
					},
				},
				PriceFilter: &schema.PriceFilter{
					StartUsageAmount: strPtr("0"),
				},
			},
		},
	}
}
