package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getSSMParameterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_ssm_parameter",
		RFunc: NewSsmParameter,
	}
}
func NewSsmParameter(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.SsmParameter{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	if !d.IsEmpty("tier") {
		r.Tier = strPtr(d.Get("tier").String())
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
