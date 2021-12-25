package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func GetNewEKSFargateProfileItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_eks_fargate_profile",
		RFunc: NewNewEKSFargateProfileItem,
	}
}
func NewNewEKSFargateProfileItem(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.NewEKSFargateProfileItem{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
