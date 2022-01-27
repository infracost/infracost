package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getVPCEndpointRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_vpc_endpoint",
		RFunc: NewVPCEndpoint,
	}
}

func NewVPCEndpoint(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	subnetIDs := d.Get("subnet_ids").Array()

	interfaces := int64(1)
	if len(subnetIDs) > 0 {
		interfaces = int64(len(subnetIDs))
	}

	r := &aws.VPCEndpoint{
		Address:    d.Address,
		Region:     d.Get("region").String(),
		Interfaces: intPtr(interfaces),
		Type:       d.Get("vpc_endpoint_type").String(),
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
