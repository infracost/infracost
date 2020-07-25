package aws

import (
	"infracost/pkg/resource"
)

func NewNatGateway(address string, region string, rawValues map[string]interface{}) resource.Resource {
	r := resource.NewBaseResource(address, rawValues, true)

	hours := resource.NewBasePriceComponent("Hours", r, "hour", "hour")
	hours.AddFilters(regionFilters(region))
	hours.AddFilters([]resource.Filter{
		{Key: "servicecode", Value: "AmazonEC2"},
		{Key: "productFamily", Value: "NAT Gateway"},
		{Key: "usagetype", Value: "/NatGateway-Hours/", Operation: "REGEX"},
	})
	r.AddPriceComponent(hours)

	return r
}
