package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getGlobalacceleratorEndpointGroupRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_globalaccelerator_endpoint_group",
		RFunc: newGlobalacceleratorEndpointGroup,
	}
}

func newGlobalacceleratorEndpointGroup(d *schema.ResourceData) schema.CoreResource {
	region := d.Get("endpointGroupRegion").String()
	r := &aws.GlobalacceleratorEndpointGroup{
		Address: d.Address,
		Region:  region,
	}

	return r
}
