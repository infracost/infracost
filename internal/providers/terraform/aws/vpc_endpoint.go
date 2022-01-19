package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getVpcEndpointRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_vpc_endpoint",
		RFunc: NewVpcEndpoint,
	}
}
func NewVpcEndpoint(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	subnetIds := d.Get("subnet_ids").Array()

	vpcEndpointInterfaces := int64(1)
	if len(subnetIds) > 0 {
		vpcEndpointInterfaces = int64(len(subnetIds))
	}

	r := &aws.VpcEndpoint{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String()), VpcEndpointInterfaces: intPtr(vpcEndpointInterfaces)}
	if !d.IsEmpty("vpc_endpoint_type") {
		r.VpcEndpointType = strPtr(d.Get("vpc_endpoint_type").String())
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
