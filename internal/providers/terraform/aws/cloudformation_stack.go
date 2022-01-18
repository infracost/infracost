package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getCloudFormationStackRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_cloudformation_stack",
		RFunc: NewCloudformationStack,
	}
}
func NewCloudformationStack(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.CloudformationStack{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	if !d.IsEmpty("template_body") {
		r.TemplateBody = strPtr(d.Get("template_body").String())
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
