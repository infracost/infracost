package aws

import (
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

func GetEC2TrafficMirroSessionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_ec2_traffic_mirror_session",
		RFunc: NewEC2TrafficMirrorSession,
	}
}

func NewEC2TrafficMirrorSession(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	region := d.Get("region").String()

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:           "Traffic mirror",
				Unit:           "hours",
				UnitMultiplier: 1,
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
		},
	}
}
