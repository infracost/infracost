package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type EC2TrafficMirrorSession struct {
	Address string
	Region  string
}

func (r *EC2TrafficMirrorSession) CoreType() string {
	return "EC2TrafficMirrorSession"
}

func (r *EC2TrafficMirrorSession) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *EC2TrafficMirrorSession) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *EC2TrafficMirrorSession) BuildResource() *schema.Resource {
	return &schema.Resource{
		Name: r.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:           "Traffic mirror",
				Unit:           "hours",
				UnitMultiplier: decimal.NewFromInt(1),
				HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
				ProductFilter: &schema.ProductFilter{
					VendorName: strPtr("aws"),
					Region:     strPtr(r.Region),
					Service:    strPtr("AmazonVPC"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/ENI-Mirror/")},
					},
				},
			},
		},
		UsageSchema: r.UsageSchema(),
	}
}
