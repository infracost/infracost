package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"
)

func GetSSMParameterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_ssm_parameter",
		RFunc: NewSSMParameter,
	}
}
func NewSSMParameter(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.SSMParameter{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	if d.Get("tier").Exists() && d.Get("tier").Type != gjson.Null {
		r.Tier = strPtr(d.Get("tier").String())
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
