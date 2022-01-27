package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getECRRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_ecr_repository",
		RFunc: NewECRRepository,
	}
}
func NewECRRepository(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.ECRRepository{
		Address: d.Address,
		Region:  d.Get("region").String(),
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
