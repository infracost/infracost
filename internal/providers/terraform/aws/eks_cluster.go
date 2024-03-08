package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getNewEKSClusterItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_eks_cluster",
		CoreRFunc: NewEKSCluster,
	}
}
func NewEKSCluster(d *schema.ResourceData) schema.CoreResource {
	r := &aws.EKSCluster{
		Address: d.Address,
		Region:  d.Get("region").String(),
		Version: d.Get("version").String(),
	}
	return r
}
