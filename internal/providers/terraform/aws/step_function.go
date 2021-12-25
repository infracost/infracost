package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"
)

func GetStepFunctionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_sfn_state_machine",
		RFunc: NewStepFunction,
	}
}
func NewStepFunction(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.StepFunction{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	if d.Get("type").Exists() && d.Get("type").Type != gjson.Null {
		r.Type = strPtr(d.Get("type").String())
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
