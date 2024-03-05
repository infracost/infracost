package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getCodeBuildProjectRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_codebuild_project",
		CoreRFunc: NewCodeBuildProject,
	}
}
func NewCodeBuildProject(d *schema.ResourceData) schema.CoreResource {
	r := &aws.CodeBuildProject{
		Address:         d.Address,
		Region:          d.Get("region").String(),
		ComputeType:     d.Get("environment.0.compute_type").String(),
		EnvironmentType: d.Get("environment.0.type").String(),
	}
	return r
}
