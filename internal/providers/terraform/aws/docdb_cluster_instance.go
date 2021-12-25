package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func GetDocDBClusterInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_docdb_cluster_instance",
		RFunc: NewDocDBClusterInstance,
	}
}
func NewDocDBClusterInstance(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.DocDBClusterInstance{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String()), InstanceClass: strPtr(d.Get("instance_class").String())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
