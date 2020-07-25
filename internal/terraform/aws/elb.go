package aws

import (
	"infracost/pkg/resource"
)

func NewElb(address string, region string, rawValues map[string]interface{}, isClassic bool) resource.Resource {
	r := resource.NewBaseResource(address, rawValues, true)

	productFamily := "Load Balancer"
	if !isClassic {
		if rawValues["load_balancer_type"] != nil && rawValues["load_balancer_type"].(string) == "network" {
			productFamily = "Load Balancer-Network"
		} else {
			productFamily = "Load Balancer-Application"
		}
	}

	hours := resource.NewBasePriceComponent("Hours", r, "hour", "hour")
	hours.AddFilters(regionFilters(region))
	hours.AddFilters([]resource.Filter{
		{Key: "servicecode", Value: "AWSELB"},
		{Key: "usagetype", Value: "/LoadBalancerUsage/", Operation: "REGEX"},
		{Key: "productFamily", Value: productFamily},
	})
	r.AddPriceComponent(hours)

	return r
}
