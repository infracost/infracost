package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getCloudfrontFunctionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_cloudfront_function",
		RFunc: newCloudfrontFunction,
	}
}

func newCloudfrontFunction(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	
	a := &aws.CloudfrontFunction{
		Address: d.Address,
		Region:  region,
	}
	
	a.PopulateUsage(u)
	
	return a.BuildResource()
}