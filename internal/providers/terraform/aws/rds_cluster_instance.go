package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getRDSClusterInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_rds_cluster_instance",
		RFunc: NewRDSClusterInstance,
	}
}

func NewRDSClusterInstance(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.RDSClusterInstance{
		Address:       d.Address,
		Region:        d.Get("region").String(),
		InstanceClass: d.Get("instance_class").String(),
		Engine:        d.Get("engine").String(),
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
