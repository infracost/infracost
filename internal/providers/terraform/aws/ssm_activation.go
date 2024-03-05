package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getSSMActivationRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_ssm_activation",
		CoreRFunc: NewSSMActivation,
	}
}

func NewSSMActivation(d *schema.ResourceData) schema.CoreResource {
	r := &aws.SSMActivation{
		Address:           d.Address,
		Region:            d.Get("region").String(),
		RegistrationLimit: d.Get("registration_limit").Int(),
	}
	return r
}
