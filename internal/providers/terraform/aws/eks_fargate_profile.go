package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getNewEKSFargateProfileItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_eks_fargate_profile",
		RFunc: NewEksFargateProfile,
	}
}
func NewEksFargateProfile(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.EksFargateProfile{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
