package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func GetCloudwatchLogGroupItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_cloudwatch_log_group",
		RFunc: NewCloudwatchLogGroupItem,
	}
}
func NewCloudwatchLogGroupItem(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.CloudwatchLogGroupItem{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
