package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getDocDBClusterInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_docdb_cluster_instance",
		CoreRFunc: NewDocDBClusterInstance,
	}
}
func NewDocDBClusterInstance(d *schema.ResourceData) schema.CoreResource {
	r := &aws.DocDBClusterInstance{
		Address:       d.Address,
		Region:        d.Get("region").String(),
		InstanceClass: d.Get("instance_class").String(),
	}
	return r
}
