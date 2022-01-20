package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getCloudFormationStackRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_cloudformation_stack",
		RFunc: NewCloudFormationStackSet,
	}
}
func NewCloudFormationStack(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.CloudFormationStack{
		Address:      d.Address,
		Region:       d.Get("region").String(),
		TemplateBody: d.Get("template_body").String(),
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
