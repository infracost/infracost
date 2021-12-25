package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func GetLightsailInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_lightsail_instance",
		RFunc: NewLightsailInstance,
	}
}
func NewLightsailInstance(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.LightsailInstance{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String()), BundleId: strPtr(d.Get("bundle_id").String())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
