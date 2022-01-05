package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type Eip struct {
	Address               *string
	CustomerOwnedIpv4Pool *string
	Instance              *string
	NetworkInterface      *string
	Region                *string
}

var EipUsageSchema = []*schema.UsageItem{}

func (r *Eip) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *Eip) BuildResource() *schema.Resource {

	if (r.CustomerOwnedIpv4Pool != nil && *r.CustomerOwnedIpv4Pool != "") || r.Instance != nil || r.NetworkInterface != nil {
		return &schema.Resource{
			Name:      *r.Address,
			NoPrice:   true,
			IsSkipped: true, UsageSchema: EipUsageSchema,
		}
	}

	region := *r.Region

	return &schema.Resource{
		Name: *r.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:           "IP address (if unused)",
				Unit:           "hours",
				UnitMultiplier: decimal.NewFromInt(1),
				HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AmazonEC2"),
					ProductFamily: strPtr("IP Address"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/ElasticIP:IdleAddress/")},
					},
				},
				PriceFilter: &schema.PriceFilter{
					StartUsageAmount: strPtr("1"),
				},
			},
		}, UsageSchema: EipUsageSchema,
	}
}
