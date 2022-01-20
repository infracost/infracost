package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getDocDBClusterInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_docdb_cluster_instance",
		RFunc: NewDocDBClusterInstance,
	}
}
func NewDocDBClusterInstance(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.DocDBClusterInstance{
		Address:       d.Address,
		Region:        d.Get("region").String(),
		InstanceClass: d.Get("instance_class").String(),
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
