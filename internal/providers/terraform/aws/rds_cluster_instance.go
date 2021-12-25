package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func GetRDSClusterInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_rds_cluster_instance",
		RFunc: NewRDSClusterInstance,
	}
}
func NewRDSClusterInstance(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.RDSClusterInstance{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String()), InstanceClass: strPtr(d.Get("instance_class").String()), Engine: strPtr(d.Get("engine").String())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
