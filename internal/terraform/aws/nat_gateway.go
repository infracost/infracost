package aws

import (
	"infracost/pkg/base"
)

func NewNatGateway(address string, region string, rawValues map[string]interface{}) base.Resource {
	r := base.NewBaseResource(address, rawValues, true)

	hours := base.NewBasePriceComponent("Hours", r, "hour", "hour")
	hours.AddFilters(regionFilters(region))
	hours.AddFilters([]base.Filter{
		{Key: "servicecode", Value: "AmazonEC2"},
		{Key: "productFamily", Value: "NAT Gateway"},
		{Key: "usagetype", Value: "/NatGateway-Hours/", Operation: "REGEX"},
	})
	r.AddPriceComponent(hours)

	return r
}
