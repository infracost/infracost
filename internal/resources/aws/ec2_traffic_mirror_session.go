package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type EC2TrafficMirroSession struct {
	Address *string
	Region  *string
}

var EC2TrafficMirroSessionUsageSchema = []*schema.UsageItem{}

func (r *EC2TrafficMirroSession) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *EC2TrafficMirroSession) BuildResource() *schema.Resource {
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
		}, UsageSchema: EC2TrafficMirroSessionUsageSchema,
	}
}
