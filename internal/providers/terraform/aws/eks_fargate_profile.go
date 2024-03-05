package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getNewEKSFargateProfileItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_eks_fargate_profile",
		CoreRFunc: NewEKSFargateProfile,
	}
}
func NewEKSFargateProfile(d *schema.ResourceData) schema.CoreResource {
	r := &aws.EKSFargateProfile{
		Address: d.Address,
		Region:  d.Get("region").String(),
	}
	return r
}
