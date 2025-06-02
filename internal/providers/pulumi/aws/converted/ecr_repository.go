package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getECRRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_ecr_repository",
		RFunc:           NewECRRepository,
		ReferenceAttributes: []string{"awsEcrLifecyclePolicy.repository"},
	}
}
func NewECRRepository(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	aws.PopulateUsage(u)
	return aws.BuildResource()
		Address: d.Address,
		Region:  d.Get("region").String(),
	}
}
