package aws

import (
	"plancosts/pkg/base"
)

type NatGatewayHours struct {
	*BaseAwsPriceComponent
}

func NewNatGatewayHours(name string, resource *NatGateway) *NatGatewayHours {
	c := &NatGatewayHours{
		NewBaseAwsPriceComponent(name, resource.BaseAwsResource, "hour"),
	}

	c.defaultFilters = []base.Filter{
		{Key: "servicecode", Value: "AmazonEC2"},
		{Key: "productFamily", Value: "NAT Gateway"},
		{Key: "usagetype", Value: "/NatGateway-Hours/", Operation: "REGEX"},
	}

	return c
}

type NatGateway struct {
	*BaseAwsResource
}

func NewNatGateway(address string, region string, rawValues map[string]interface{}) *NatGateway {
	r := &NatGateway{
		BaseAwsResource: NewBaseAwsResource(address, region, rawValues),
	}
	r.BaseAwsResource.priceComponents = []base.PriceComponent{
		NewNatGatewayHours("Hours", r),
	}
	return r
}
