package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getCloudFormationStackSetRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_cloudformation_stack_set",
		RFunc: NewCloudformationStackSet,
	}
}
func NewCloudformationStackSet(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.CloudformationStackSet{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	if !d.IsEmpty("template_body") {
		r.TemplateBody = strPtr(d.Get("template_body").String())
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
