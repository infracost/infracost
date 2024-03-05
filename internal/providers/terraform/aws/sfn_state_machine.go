package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getStepFunctionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_sfn_state_machine",
		CoreRFunc: NewSFnStateMachine,
	}
}

func NewSFnStateMachine(d *schema.ResourceData) schema.CoreResource {
	r := &aws.SFnStateMachine{
		Address: d.Address,
		Region:  d.Get("region").String(),
		Type:    d.Get("type").String(),
	}
	return r
}
