package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getLightsailInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_lightsail_instance",
		RFunc: NewLightsailInstance,
	}
}

func NewLightsailInstance(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.LightsailInstance{
		Address:  d.Address,
		BundleID: d.Get("bundle_id").String(),
		Region:   d.Get("region").String(),
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
