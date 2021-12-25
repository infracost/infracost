package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func GetNewEKSClusterItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_eks_cluster",
		RFunc: NewNewEKSClusterItem,
	}
}
func NewNewEKSClusterItem(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.NewEKSClusterItem{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
