package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func GetECRRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_ecr_repository",
		RFunc: NewECR,
	}
}
func NewECR(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.ECR{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
