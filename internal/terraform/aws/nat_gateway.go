package aws

import (
	"infracost/pkg/resource"
)

func NewNatGateway(address string, region string, rawValues map[string]interface{}) resource.Resource {
	r := resource.NewBaseResource(address, rawValues, true)

	hoursProductFilter := &resource.ProductFilter{
		VendorName:    strPtr("aws"),
		Region:        strPtr(region),
		Service:       strPtr("AmazonEC2"),
		ProductFamily: strPtr("NAT Gateway"),
		AttributeFilters: &[]resource.AttributeFilter{
			{Key: "usagetype", ValueRegex: strPtr("/NatGateway-Hours/")},
		},
	}
	hours := resource.NewBasePriceComponent("hours", r, "hour", "hour", hoursProductFilter, nil)
	r.AddPriceComponent(hours)

	return r
}
