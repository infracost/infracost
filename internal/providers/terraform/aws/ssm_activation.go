package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getSSMActivationRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_ssm_activation",
		RFunc: NewSsmActivation,
	}
}
func NewSsmActivation(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.SsmActivation{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	if !d.IsEmpty("registration_limit") {
		r.RegistrationLimit = intPtr(d.Get("registration_limit").Int())
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
