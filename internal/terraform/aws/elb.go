package aws

import (
	"fmt"
	"infracost/pkg/base"
)

type ElbHours struct {
	*BaseAwsPriceComponent
}

func NewElbHours(name string, resource *Elb, isClassic bool) *ElbHours {
	c := &ElbHours{
		NewBaseAwsPriceComponent(name, resource.BaseAwsResource, "hour"),
	}

	defaultProductFamily := "Load Balancer"
	if !isClassic {
		defaultProductFamily = "Load Balancer-Application"
	}

	c.defaultFilters = []base.Filter{
		{Key: "servicecode", Value: "AWSELB"},
		{Key: "productFamily", Value: defaultProductFamily},
		{Key: "usagetype", Value: "/LoadBalancerUsage/", Operation: "REGEX"},
	}

	if !isClassic {
		c.valueMappings = []base.ValueMapping{
			{
				FromKey: "load_balancer_type",
				ToKey:   "productFamily",
				ToValueFn: func(fromValue interface{}) string {
					if fmt.Sprintf("%v", fromValue) == "network" {
						return "Load Balancer-Network"
					}
					return "Load Balancer-Application"
				},
			},
		}
	}

	return c
}

type Elb struct {
	*BaseAwsResource
}

func NewElb(address string, region string, rawValues map[string]interface{}, isClassic bool) *Elb {
	r := &Elb{
		BaseAwsResource: NewBaseAwsResource(address, region, rawValues),
	}
	r.BaseAwsResource.priceComponents = []base.PriceComponent{
		NewElbHours("Hours", r, isClassic),
	}
	return r
}
