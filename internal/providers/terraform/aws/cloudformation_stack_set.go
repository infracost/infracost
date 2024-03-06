package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getCloudFormationStackSetRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_cloudformation_stack_set",
		CoreRFunc: NewCloudFormationStackSet,
	}
}
func NewCloudFormationStackSet(d *schema.ResourceData) schema.CoreResource {
	r := &aws.CloudFormationStackSet{
		Address:      d.Address,
		Region:       d.Get("region").String(),
		TemplateBody: d.Get("template_body").String(),
	}
	return r
}
