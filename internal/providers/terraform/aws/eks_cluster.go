package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getNewEKSClusterItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_eks_cluster",
		RFunc: NewEKSCluster,
	}
}
func NewEKSCluster(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.EKSCluster{
		Address: d.Address,
		Region:  d.Get("region").String(),
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
