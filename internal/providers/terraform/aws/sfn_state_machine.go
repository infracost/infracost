package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getStepFunctionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_sfn_state_machine",
		RFunc: NewSFnStateMachine,
	}
}

func NewSFnStateMachine(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.SFnStateMachine{
		Address: d.Address,
		Region:  d.Get("region").String(),
		Type:    d.Get("type").String(),
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
