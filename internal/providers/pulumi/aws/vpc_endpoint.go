package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getVPCEndpointRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_vpc_endpoint",
		RFunc: NewVPCEndpoint,
		ReferenceAttributes: []string{
			"subnetIds",
			"vpcId",
		},
	}
}

func NewVPCEndpoint(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	subnetIDs := len(d.Get("subnetIds").Array())

	// if the length of the subnetIds attribute is zero this means that the attribute
	// has been modified with a subnet id that is yet to exist. In this instance we'll
	// use the reference attribute instead. In most cases this should have the accurate
	// number of subnet_ids.
	if subnetIDs == 0 {
		subnetIDs = len(d.References("subnetIds"))
	}

	var interfaces int64 = 1
	if subnetIDs > 0 {
		interfaces = int64(subnetIDs)
	}

	r := &aws.VPCEndpoint{
		Address:    d.Address,
		Region:     d.Get("region").String(),
		Interfaces: intPtr(interfaces),
		Type:       d.Get("vpcEndpointType").String(),
	}
	
	r.PopulateUsage(u)
	
	return r.BuildResource()
}