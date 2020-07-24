package aws

import (
	"infracost/pkg/base"
)

func NewElb(address string, region string, rawValues map[string]interface{}, isClassic bool) base.Resource {
	r := base.NewBaseResource(address, rawValues, true)

	productFamily := "Load Balancer"
	if !isClassic {
		if rawValues["load_balancer_type"] != nil && rawValues["load_balancer_type"].(string) == "network" {
			productFamily = "Load Balancer-Network"
		} else {
			productFamily = "Load Balancer-Application"
		}
	}

	hours := base.NewBasePriceComponent("Hours", r, "hour", "hour")
	hours.AddFilters(regionFilters(region))
	hours.AddFilters([]base.Filter{
		{Key: "servicecode", Value: "AWSELB"},
		{Key: "usagetype", Value: "/LoadBalancerUsage/", Operation: "REGEX"},
		{Key: "productFamily", Value: productFamily},
	})
	r.AddPriceComponent(hours)

	return r
}
