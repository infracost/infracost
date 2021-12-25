package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"
)

func GetSSMActivationRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_ssm_activation",
		RFunc: NewSSMActivation,
	}
}
func NewSSMActivation(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.SSMActivation{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	if d.Get("registration_limit").Exists() && d.Get("registration_limit").Type != gjson.Null {
		r.RegistrationLimit = intPtr(d.Get("registration_limit").Int())
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
