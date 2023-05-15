package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func GetNewEKSFargateProfileItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "states.aws.eks.fargate_profile.present",
		RFunc: NewEKSFargateProfile,
	}
}
func NewEKSFargateProfile(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.EKSFargateProfile{
		Address: d.Address,
		Region:  d.Get("region").String(),
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
