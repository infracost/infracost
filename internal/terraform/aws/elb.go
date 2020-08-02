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

	hoursProductFilter := &resource.ProductFilter{
		VendorName:    strPtr("aws"),
		Region:        strPtr(region),
		Service:       strPtr("AWSELB"),
		ProductFamily: strPtr(productFamily),
		AttributeFilters: &[]resource.AttributeFilter{
			{Key: "usagetype", ValueRegex: strPtr("/LoadBalancerUsage/")},
		},
	}
	hours := resource.NewBasePriceComponent("Hours", r, "hour", "hour", hoursProductFilter, nil)
	r.AddPriceComponent(hours)

	return r
}
