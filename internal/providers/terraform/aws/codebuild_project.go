package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getCodebuildProjectRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_codebuild_project",
		RFunc: NewCodebuildProject,
	}
}
func NewCodebuildProject(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.CodebuildProject{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String()), Environment0ComputeType: strPtr(d.Get("environment.0.compute_type").String()), Environment0Type: strPtr(d.Get("environment.0.type").String())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
