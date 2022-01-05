package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getECRRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_ecr_repository",
		RFunc: NewEcrRepository,
	}
}
func NewEcrRepository(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.EcrRepository{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
