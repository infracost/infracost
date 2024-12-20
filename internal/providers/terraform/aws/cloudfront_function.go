package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getCloudfrontFunctionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_cloudfront_function",
		CoreRFunc: newCloudfrontFunction,
	}
}

func newCloudfrontFunction(d *schema.ResourceData) schema.CoreResource {
	region := d.Region
	return &aws.CloudfrontFunction{
		Address: d.Address,
		Region:  region,
	}
}
