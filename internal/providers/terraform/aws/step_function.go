package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getStepFunctionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_sfn_state_machine",
		RFunc: NewSfnStateMachine,
	}
}
func NewSfnStateMachine(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.SfnStateMachine{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	if !d.IsEmpty("type") {
		r.Type = strPtr(d.Get("type").String())
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
