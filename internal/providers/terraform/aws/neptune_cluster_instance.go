package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getNeptuneClusterInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_neptune_cluster_instance",
		RFunc: NewNeptuneClusterInstance,
	}
}

func NewNeptuneClusterInstance(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.NeptuneClusterInstance{
		Address:       d.Address,
		Region:        d.Get("region").String(),
		InstanceClass: d.Get("instance_class").String(),
	}

	if !d.IsEmpty("count") {
		r.Count = intPtr(d.Get("count").Int())
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
