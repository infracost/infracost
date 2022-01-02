package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type Ec2TrafficMirrorSession struct {
	Address *string
	Region  *string
}

var Ec2TrafficMirrorSessionUsageSchema = []*schema.UsageItem{}

func (r *Ec2TrafficMirrorSession) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *Ec2TrafficMirrorSession) BuildResource() *schema.Resource {
	region := *r.Region

	return &schema.Resource{
		Name: *r.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:           "Traffic mirror",
				Unit:           "hours",
				UnitMultiplier: decimal.NewFromInt(1),
				HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
				ProductFilter: &schema.ProductFilter{
					VendorName: strPtr("aws"),
					Region:     strPtr(region),
					Service:    strPtr("AmazonVPC"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/ENI-Mirror/")},
					},
				},
			},
		}, UsageSchema: Ec2TrafficMirrorSessionUsageSchema,
	}
}
