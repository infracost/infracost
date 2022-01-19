package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getRoute53ResolverEndpointRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_route53_resolver_endpoint",
		RFunc: NewRoute53ResolverEndpoint,
	}
}
func NewRoute53ResolverEndpoint(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	resolverEndpoints := int64(len(d.Get("ip_address").Array()))

	r := &aws.Route53ResolverEndpoint{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String()), ResolverEndpoints: intPtr(resolverEndpoints)}
	r.PopulateUsage(u)
	return r.BuildResource()
}
