package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getGlobalAcceleratorRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_globalaccelerator_accelerator",
		RFunc: newGlobalAccelerator,
	}
}

func newGlobalAccelerator(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	name := d.Get("name").String()
	ipAddressType := d.Get("ip_address_type").String()
	enabled := d.Get("enabled").Bool()

	r := &aws.GlobalAccelerator{
		Name:          name,
		IPAddressType: ipAddressType,
		Enabled:       enabled,
	}
	r.PopulateUsage(u)

	return r.BuildResource()
}
